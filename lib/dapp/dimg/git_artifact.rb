module Dapp
  module Dimg
    # Git repo artifact
    class GitArtifact
      include Helper::Tar
      include Helper::Trivia

      attr_reader :repo
      attr_reader :name
      attr_reader :as

      # FIXME: переименовать cwd в from

      # rubocop:disable Metrics/ParameterLists
      def initialize(repo, dimg, to:, name: nil, branch: nil, commit: nil,
                     cwd: nil, include_paths: nil, exclude_paths: nil, owner: nil, group: nil, as: nil,
                     stages_dependencies: {})
        @repo = repo
        @dimg = dimg
        @name = name

        @branch = branch || repo.dapp.options[:git_artifact_branch] || repo.branch
        @commit = commit

        @to = to
        @cwd = (cwd.nil? || cwd.empty? || cwd == '/') ? '' : File.expand_path(File.join('/', cwd, '/'))[1..-1]
        @include_paths = include_paths
        @exclude_paths = exclude_paths
        @owner = owner
        @group = group
        @as = as

        @stages_dependencies = stages_dependencies
      end
      # rubocop:enable Metrics/ParameterLists

      def embedded_artifacts
        [submodules_artifacts, nested_git_directory_artifacts].flatten
      end

      def submodules_artifacts
        commit = dev_mode? ? nil : latest_commit
        repo.submodules_params(commit,
                               paths: include_paths_or_cwd,
                               exclude_paths: exclude_paths(true)).map(&method(:submodule_artifact))
      end

      def submodule_artifact(submodule_params)
        embedded_artifact(submodule_params)
      rescue Rugged::InvalidError => e
        raise Error::Rugged, code: :git_local_incorrect_gitmodules_params, data: { error: e.message }
      end

      def nested_git_directory_artifacts
        return [] unless dev_mode?
        repo
          .nested_git_directories_patches(paths: include_paths_or_cwd, exclude_paths: exclude_paths(true), **diff_patches_options)
          .map(&method(:nested_git_directory_artifact))
      end

      def nested_git_directory_artifact(patch)
        params = {}.tap do |p|
          p[:path] = patch.delta.new_file[:path]
          p[:type] = :local
        end
        embedded_artifact(params)
      end

      def embedded_artifact(embedded_params)
        embedded_rel_path = embedded_params[:path]
        embedded_repo     = begin
          if embedded_params[:type] == :remote
            GitRepo::Remote.get_or_create(repo.dapp, embedded_rel_path,
                                          url: embedded_params[:url],
                                          branch: embedded_params[:branch],
                                          ignore_git_fetch: dimg.ignore_git_fetch )
          elsif embedded_params[:type] == :local
            embedded_path = File.join(repo.workdir_path, embedded_params[:path])
            GitRepo::Local.new(repo.dapp, embedded_rel_path, embedded_path)
          else
            raise
          end
        end

        self.class.new(embedded_repo, dimg, embedded_artifact_options(embedded_params))
      end

      def embedded_artifact_options(embedded_params)
        embedded_rel_path = embedded_params[:path]

        {}.tap do |options|
          options[:name]                = repo.dapp.consistent_uniq_slugify("embedded-#{embedded_rel_path}")
          options[:cwd]                 = embedded_inherit_path(cwd, embedded_rel_path).last
          options[:to]                  = Pathname(cwd).subpath_of?(embedded_rel_path) ? to : File.join(to, embedded_rel_path)
          options[:include_paths]       = embedded_inherit_paths(include_paths, embedded_rel_path)
          options[:exclude_paths]       = embedded_inherit_paths(exclude_paths, embedded_rel_path)
          options[:stages_dependencies] = begin
            stages_dependencies
              .map { |stage, paths| [stage, embedded_inherit_paths(paths, embedded_rel_path)] }
              .to_h
          end
          options[:branch]              = embedded_params[:branch]
          options[:owner]               = owner
          options[:group]               = group
        end
      end

      def embedded_inherit_paths(paths, embedded_rel_path)
        paths
          .select { |path| check_path?(embedded_rel_path, path) || check_subpath?(embedded_rel_path, path) }
          .map { |path| embedded_inherit_path(path, embedded_rel_path) }
          .flatten
          .compact
      end

      def embedded_inherit_path(path, embedded_rel_path)
        path_parts      = path.split('/')
        test_path       = nil
        inherited_paths = []
        until path_parts.empty?
          last_part_path = path_parts.shift
          test_path      = [test_path, last_part_path].compact.join('/')

          non_match    = !File.fnmatch(test_path, embedded_rel_path, File::FNM_PATHNAME|File::FNM_DOTMATCH)
          part_for_all = (last_part_path == '**')

          if non_match || part_for_all
            inherited_paths << [last_part_path, path_parts].flatten.join('/')
            break unless part_for_all
          end
        end

        inherited_paths
      end

      def cwd_type(stage)
        if dev_mode?
          p = repo.workdir_path.join(cwd)
          if p.exist?
            if p.directory?
              :directory
            else
              :file
            end
          end
        elsif cwd == ''
          :directory
        else
          commit = repo.lookup_commit(stage.layer_commit(self))

          cwd_entry = begin
            commit.tree.path(cwd)
          rescue Rugged::TreeError
          end

          if cwd_entry
            if cwd_entry[:type] == :tree
              :directory
            else
              :file
            end
          end
        end
      end

      def apply_archive_command(stage)
        [].tap do |commands|
          if archive_any_changes?(stage)
            case cwd_type(stage)
            when :directory
              stage.image.add_service_change_label(repo.dapp.dimgstage_g_a_type_label(paramshash).to_sym => 'directory')

              commands << "#{repo.dapp.install_bin} #{credentials.join(' ')} -d \"#{to}\""
              commands << "#{sudo}#{repo.dapp.tar_bin} -xf #{archive_file(stage)} -C \"#{to}\""
            when :file
              stage.image.add_service_change_label(repo.dapp.dimgstage_g_a_type_label(paramshash).to_sym => 'file')

              commands << "#{repo.dapp.install_bin} #{credentials.join(' ')} -d \"#{File.dirname(to)}\""
              commands << "#{sudo}#{repo.dapp.tar_bin} -xf #{archive_file(stage)} -C \"#{File.dirname(to)}\""
            end
          end
        end
      end

      def archive_type(stage)
        stage.prev_stage.image.labels[repo.dapp.dimgstage_g_a_type_label(paramshash)].to_s.to_sym
      end

      def apply_patch_command(stage)
        [].tap do |commands|
          if dev_mode?
            if any_changes?(*dev_patch_stage_commits(stage))
              case archive_type(stage)
              when :directory
                files_to_remove_file_name = file_name('dev_files_to_remove')
                File.open(dimg.tmp_path('archives', files_to_remove_file_name), File::RDWR | File::CREAT) do |f|
                  diff_patches(*dev_patch_stage_commits(stage))
                    .map {|p| File.join(to, cwd, p.delta.new_file[:path])}
                    .each(&f.method(:puts))
                end

                commands << "#{repo.dapp.rm_bin} -rf $(#{repo.dapp.cat_bin} #{dimg.container_tmp_path('archives', files_to_remove_file_name)})"
                commands << "#{repo.dapp.install_bin} #{credentials.join(' ')} -d \"#{to}\""
                commands << "#{sudo}#{repo.dapp.tar_bin} -xf #{archive_file(stage)} -C \"#{to}\""
                commands << "#{repo.dapp.find_bin} \"#{to}\" -empty -type d -delete"
              when :file
                commands << "#{repo.dapp.rm_bin} -rf \"#{to}\""
                commands << "#{repo.dapp.install_bin} #{credentials.join(' ')} -d \"#{File.dirname(to)}\""
                commands << "#{sudo}#{repo.dapp.tar_bin} -xf #{archive_file(stage)} -C \"#{File.dirname(to)}\""
              else
                raise
              end
            end
          else
            if patch_any_changes?(stage)
              case archive_type(stage)
              when :directory
                commands << "#{repo.dapp.install_bin} #{credentials.join(' ')} -d \"#{to}\""
                commands << "#{sudo}#{repo.dapp.git_bin} apply --whitespace=nowarn --directory=\"#{to}\" --unsafe-paths #{patch_file(stage, *patch_stage_commits(stage))}"
              when :file
                commands << "#{repo.dapp.install_bin} #{credentials.join(' ')} -d \"#{File.dirname(to)}\""
                commands << "#{sudo}#{repo.dapp.git_bin} apply --whitespace=nowarn --directory=\"#{File.dirname(to)}\" --unsafe-paths #{patch_file(stage, *patch_stage_commits(stage))}"
              else
                raise
              end
            end
          end
        end
      end

      def stage_dependencies_checksum(stage)
        return [] if (stage_dependencies = stages_dependencies[stage.name]).empty?

        paths = base_paths(stage_dependencies, true)
        commit = dev_mode? ? nil : latest_commit

        stage_dependencies_key = [stage.name, commit]

        @stage_dependencies_checksums ||= {}
        @stage_dependencies_checksums[stage_dependencies_key] ||= begin
          if dev_mode?
            dev_patch_hash(paths: paths)
          else
            if (entries = repo_entries(commit, paths: paths)).empty?
              repo.dapp.log_warning(desc: { code: :stage_dependencies_not_found,
                                            data: { repo: repo.respond_to?(:url) ? repo.url : 'local',
                                                    dependencies: stage_dependencies.join(', ') } })
            end

            entries
              .sort_by {|root, entry| File.join(root, entry[:name])}
              .reduce(nil) {|prev_hash, (root, entry)|
                content = nil
                content = repo.lookup_object(entry[:oid]).content if entry[:type] == :blob

                hexdigest prev_hash, File.join(root, entry[:name]), entry[:filemode].to_s, content
              }
          end
        end
      end

      def patch_size(from_commit)
        to_commit = dev_mode? ? nil : latest_commit
        diff_patches(from_commit, to_commit).reduce(0) do |bytes, patch|
          bytes += patch.delta.deleted? ? patch.delta.old_file[:size] : patch.delta.new_file[:size]
          bytes
        end
      end

      def dev_patch_hash(**options)
        return unless dev_mode?
        hexdigest begin
          diff_patches(nil, nil, **options).map do |patch|
            file = patch.delta.new_file
            [file[:path], File.read(File.join(repo.workdir_path, file[:path])), file[:mode]]
          end
        end
      end

      def latest_commit
        @latest_commit ||= (commit || repo.latest_commit(branch))
      end

      def paramshash
        hexdigest full_name, to, cwd, *include_paths, *exclude_paths, owner, group
      end

      def full_name
        "#{repo.name}#{name ? "_#{name}" : nil}"
      end

      def archive_any_changes?(stage)
        repo_entries(stage_commit(stage)).any?
      end

      def patch_any_changes?(stage)
        any_changes?(*patch_stage_commits(stage))
      end

      def empty?
        repo_entries(latest_commit).empty?
      end

      protected

      def hexdigest(*args)
        Digest::SHA256.hexdigest args.compact.map {|arg| arg.to_s.force_encoding("ASCII-8BIT")}.join(":::")
      end

      attr_reader :dimg
      attr_reader :to
      attr_reader :commit
      attr_reader :branch
      attr_reader :cwd
      attr_reader :owner
      attr_reader :group
      attr_reader :stages_dependencies

      def sudo
        repo.dapp.sudo_command(owner: owner, group: group)
      end

      def credentials
        [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact
      end

      def archive_file(stage)
        commit = stage_commit(stage)
        if repo.dapp.options[:use_system_tar]
          archive_file_with_system_tar(stage, commit)
        else
          archive_file_with_tar_writer(stage, commit)
        end
        dimg.container_tmp_path('archives', archive_file_name(commit))
      end

      def archive_file_with_tar_writer(stage, commit)
        tar_write(dimg.tmp_path('archives', archive_file_name(commit))) do |tar|
          each_archive_entry(stage, commit) do |path, content, mode|
            if mode == 0o120000 # symlink
              tar.add_symlink path, content, mode
            else
              tar.add_file path, mode do |tf|
                tf.write content
              end
            end
          end
        end
      rescue Gem::Package::TooLongFileName => e
        raise Error::TarWriter, message: e.message
      end

      def archive_file_with_system_tar(stage, commit)
        dimg.tmp_path('archives', archive_file_name(commit)).tap do |archive_path|
          relative_archive_file_path = File.join('archives_files', file_name(commit))
          each_archive_entry(stage, commit) do |path, content, mode|
            file_path = dimg.tmp_path(relative_archive_file_path, path)

            if mode == 0o120000 # symlink
              FileUtils.symlink(content, file_path)
            else
              IO.write(file_path, content)
              FileUtils.chmod(mode, file_path)
            end
          end

          repo.dapp.shellout!("tar -C #{dimg.tmp_path(relative_archive_file_path)} -cf #{archive_path} .")
        end
      end

      def slice_cwd(stage, path)
        return path if cwd.empty?

        case cwd_type(stage)
        when :directory
          path
            .reverse
            .chomp(cwd.reverse)
            .chomp('/')
            .reverse
        when :file
          File.basename(to)
        else
          raise
        end
      end

      def archive_file_name(commit)
        file_name(commit, ext: 'tar')
      end

      def patch_file(stage, from_commit, to_commit)
        File.open(dimg.tmp_path('patches', patch_file_name(from_commit, to_commit)), File::RDWR | File::CREAT) do |f|
          diff_patches(from_commit, to_commit).each { |patch| f.write change_patch_new_file_path(stage, patch) }
        end
        dimg.container_tmp_path('patches', patch_file_name(from_commit, to_commit))
      end

      # rubocop:disable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity
      def change_patch_new_file_path(stage, patch)
        patch.to_s.lines.tap do |lines|
          modify_patch_line = proc do |line_number, path_char|
            action_part, path_part = lines[line_number].strip.split(' ', 2)
            if (path_with_cwd = path_part.partition("#{path_char}/").last).start_with?(cwd)
              native_path = case archive_type(stage)
              when :directory
                path_with_cwd.sub(cwd, '')
              when :file
                File.basename(to)
              else
                raise
              end

              if native_path
                expected_path = File.join(path_char, native_path)
                lines[line_number] = [action_part, expected_path].join(' ') + "\n"
              end

              native_path
            end
          end

          modify_patch = proc do |*modify_patch_line_args|
            native_paths = modify_patch_line_args.map { |args| modify_patch_line.call(*args) }
            unless (native_paths = native_paths.compact.uniq).empty?
              raise Error::Build, code: :unsupported_patch_format, data: { patch: patch.to_s } unless native_paths.one?
              native_path = native_paths.first
              lines[0] = ['diff --git', File.join('a', native_path), File.join('b', native_path)].join(' ') + "\n"
            end
          end

          case
          when patch.delta.deleted? then modify_patch.call([3, 'a'])
          when patch.delta.added? then modify_patch.call([4, 'b'])
          when patch.delta.modified?
            if patch_file_mode_changed?(patch)
              modify_patch.call([4, 'a'], [5, 'b'])
            else
              modify_patch.call([2, 'a'], [3, 'b'])
            end
          else
            raise
          end
        end.join
      end
      # rubocop:enable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity

      def patch_file_mode_changed?(patch)
        patch.delta.old_file[:mode] != patch.delta.new_file[:mode]
      end

      def patch_file_name(from_commit, to_commit)
        file_name(from_commit, to_commit, ext: 'patch')
      end

      def file_name(*args, ext: nil)
        "#{[paramshash, args].flatten.compact.join('_')}#{".#{ext}" unless ext.nil? }"
      end

      def repo_entries(commit, paths: include_paths_or_cwd)
        (@repo_entries ||= {})[[commit, paths]] ||= begin
          repo
            .entries(commit, paths: paths, exclude_paths: exclude_paths(true))
            .select { |_, entry| !submodule_mode?(entry[:filemode]) }
        end
      end

      def each_archive_entry(stage, commit)
        if dev_mode? && stage.name != :g_a_archive
          diff_patches(commit, nil).each do |patch|
            file = patch.delta.new_file
            host_file_path = File.join(repo.workdir_path, file[:path])
            next unless File.exist?(host_file_path)

            content = File.read(host_file_path)
            yield slice_cwd(stage, file[:path]), content, file[:mode]
          end
        else
          repo_entries(commit).each do |root, entry|
            next unless entry[:type] == :blob

            entry_file_path = File.join(root, entry[:name])
            content = repo.lookup_object(entry[:oid]).content
            yield slice_cwd(stage, entry_file_path), content, entry[:filemode]
          end
        end
      end

      def diff_patches(from_commit, to_commit, paths: include_paths_or_cwd)
        (@diff_patches ||= {})[[from_commit, to_commit, paths]] ||= begin
          repo
            .patches(from_commit, to_commit, paths: paths, exclude_paths: exclude_paths(true), **diff_patches_options)
            .select do |patch|
              file_mode = patch.delta.status == :deleted ? patch.delta.old_file[:mode] : patch.delta.new_file[:mode]
              !(submodule_mode?(file_mode) || # FIXME: https://github.com/libgit2/rugged/issues/727
                nested_git_directory_mode?(file_mode))
            end
        end
      end

      def diff_patches_options
        {}.tap do |opts|
          opts[:force_text] = true
          if dev_mode?
            opts[:include_untracked] = true
            opts[:recurse_untracked_dirs] = true
          end
        end
      end

      def submodule_mode?(mode) # FIXME
        mode == 0o160000
      end

      def nested_git_directory_mode?(mode)
        mode == 0o040000
      end

      def include_paths_or_cwd
        case
        when !include_paths(true).empty? then include_paths(true)
        when !cwd.empty? then [cwd]
        else
          []
        end
      end

      def exclude_paths(with_cwd = false)
        repo.exclude_paths + base_paths(@exclude_paths, with_cwd)
      end

      def include_paths(with_cwd = false)
        base_paths(@include_paths, with_cwd)
      end

      def base_paths(paths, with_cwd = false)
        [paths].flatten.compact.map do |path|
          if with_cwd && !cwd.empty?
            File.join(cwd, path)
          else
            path
          end
            .chomp('/')
            .reverse.chomp('/')
            .reverse
        end
      end

      def stage_commit(stage)
        stage.layer_commit(self)
      end

      def patch_stage_commits(stage)
        [stage.prev_stage.layer_commit(self), stage.layer_commit(self)]
      end

      def dev_patch_stage_commits(stage)
        [stage.prev_stage.layer_commit(self), nil]
      end

      def any_changes?(from_commit, to_commit)
        diff_patches(from_commit, to_commit).any?
      end

      def dev_mode?
        local? && dimg.dev_mode?
      end

      def local?
        repo.is_a? GitRepo::Local
      end
    end
  end
end
