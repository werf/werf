module Dapp
  module Dimg
    # Git repo artifact
    class GitArtifact
      include Helper::Tar

      attr_reader :repo
      attr_reader :name

      # FIXME: переименовать cwd в from

      # rubocop:disable Metrics/ParameterLists
      def initialize(repo, to:, name: nil, branch: nil, commit: nil,
                     cwd: nil, include_paths: nil, exclude_paths: nil, owner: nil, group: nil,
                     stages_dependencies: {})
        @repo = repo
        @name = name

        @branch = branch || repo.dimg.dapp.options[:git_artifact_branch] || repo.branch
        @commit = commit

        @to = to
        @cwd = (cwd.nil? || cwd.empty? || cwd == '/') ? '' : File.expand_path(File.join('/', cwd, '/'))[1..-1]
        @include_paths = include_paths
        @exclude_paths = exclude_paths
        @owner = owner
        @group = group

        @stages_dependencies = stages_dependencies
      end
      # rubocop:enable Metrics/ParameterLists

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
              stage.image.add_service_change_label :"dapp-git-#{paramshash}-type" => 'directory'

              commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d \"#{to}\""
              commands << "#{sudo}#{repo.dimg.dapp.tar_bin} -xf #{archive_file(stage, *archive_stage_commits(stage))} -C \"#{to}\""
            when :file
              stage.image.add_service_change_label :"dapp-git-#{paramshash}-type" => 'file'

              commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d \"#{File.dirname(to)}\""
              commands << "#{sudo}#{repo.dimg.dapp.tar_bin} -xf #{archive_file(stage, *archive_stage_commits(stage))} -C \"#{File.dirname(to)}\""
            end
          end
        end
      end

      def archive_type
        repo.dimg.stage_by_name(:g_a_archive).image.labels["dapp-git-#{paramshash}-type"].to_s.to_sym
      end

      def apply_patch_command(stage)
        [].tap do |commands|
          if dev_mode?
            if any_changes?(*dev_patch_stage_commits(stage))
              case archive_type
              when :directory
                changed_files = diff_patches(*dev_patch_stage_commits(stage)).map {|p| "\"#{File.join(to, cwd, p.delta.new_file[:path])}\""}
                commands << "#{repo.dimg.dapp.rm_bin} -rf #{changed_files.join(' ')}"
                commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d \"#{to}\""
                commands << "#{sudo}#{repo.dimg.dapp.tar_bin} -xf #{archive_file(stage, *dev_patch_stage_commits(stage))} -C \"#{to}\""
              when :file
                commands << "#{repo.dimg.dapp.rm_bin} -rf \"#{to}\""
                commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d \"#{File.dirname(to)}\""
                commands << "#{sudo}#{repo.dimg.dapp.tar_bin} -xf #{archive_file(stage, *dev_patch_stage_commits(stage))} -C \"#{File.dirname(to)}\""
              else
                raise
              end
            end
          else
            if patch_any_changes?(stage)
              case archive_type
              when :directory
                commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d \"#{to}\""
                commands << "#{sudo}#{repo.dimg.dapp.git_bin} apply --whitespace=nowarn --directory=\"#{to}\" --unsafe-paths #{patch_file(stage, *patch_stage_commits(stage))}"
              when :file
                commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d \"#{File.dirname(to)}\""
                commands << "#{sudo}#{repo.dimg.dapp.git_bin} apply --whitespace=nowarn --directory=\"#{File.dirname(to)}\" --unsafe-paths #{patch_file(stage, *patch_stage_commits(stage))}"
              else
                raise
              end
            end
          end
        end
      end

      def stage_dependencies_checksum(stage)
        return [] if (stage_dependencies = stages_dependencies[stage.name]).empty?

        paths = (include_paths(true) + base_paths(stage_dependencies, true)).uniq
        to_commit = dev_mode? ? nil : latest_commit

        stage_dependencies_key = [stage.name, to_commit]
        @stage_dependencies_checksums ||= {}
        @stage_dependencies_checksums[stage_dependencies_key] = begin
          if @stage_dependencies_checksums.key?(stage_dependencies_key)
            @stage_dependencies_checksums[stage_dependencies_key]
          else
            if (patches = diff_patches(nil, to_commit, paths: paths)).empty?
              repo.dimg.dapp.log_warning(desc: { code: :stage_dependencies_not_found,
                                                 data: { repo: repo.respond_to?(:url) ? repo.url : 'local',
                                                         dependencies: stage_dependencies.join(', ') } })
            end

            patches.sort_by {|patch| patch.delta.new_file[:path]}
              .reduce(nil) {|prev_hash, patch|
              Digest::SHA256.hexdigest [
                prev_hash,
                patch.delta.new_file[:path],
                patch.delta.new_file[:mode].to_s,
                patch.to_s
              ].compact.join(':::')
            }
          end
        end
      end

      def patch_size(from_commit, to_commit)
        diff_patches(from_commit, to_commit).reduce(0) do |bytes, patch|
          patch.hunks.each do |hunk|
            hunk.lines.each do |l|
              bytes +=
                case l.line_origin
                when :eof_newline_added, :eof_newline_removed then 1
                when :addition, :deletion, :binary            then l.content.size
                else # :context, :file_header, :hunk_header, :eof_no_newline
                  0
                end
            end
          end
          bytes
        end
      end

      def dev_patch_hash(stage)
        return unless dev_mode?

        Digest::SHA256.hexdigest(diff_patches(latest_commit, nil).map do |patch|
          change_patch_new_file_path(stage, patch)
        end.join(':::'))
      end

      def latest_commit
        @latest_commit ||= (commit || repo.latest_commit(branch))
      end

      def paramshash
        Digest::SHA256.hexdigest [full_name, to, cwd, *include_paths, *exclude_paths, owner, group].map(&:to_s).join(':::')
      end

      def full_name
        "#{repo.name}#{name ? "_#{name}" : nil}"
      end

      def archive_any_changes?(stage)
        any_changes?(*archive_stage_commits(stage))
      end

      def patch_any_changes?(stage)
        any_changes?(*patch_stage_commits(stage))
      end

      protected

      attr_reader :to
      attr_reader :commit
      attr_reader :branch
      attr_reader :cwd
      attr_reader :owner
      attr_reader :group
      attr_reader :stages_dependencies

      def sudo
        repo.dimg.dapp.sudo_command(owner: owner, group: group)
      end

      def credentials
        [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact
      end

      def archive_file(stage, from_commit, to_commit)
        if repo.dimg.dapp.options[:use_system_tar]
          archive_file_with_system_tar(stage, from_commit, to_commit)
        else
          archive_file_with_tar_writer(stage, from_commit, to_commit)
        end
        repo.dimg.container_tmp_path('archives', archive_file_name(from_commit, to_commit))
      end

      def archive_file_with_tar_writer(stage, from_commit, to_commit)
        tar_write(repo.dimg.tmp_path('archives', archive_file_name(from_commit, to_commit))) do |tar|
          archive_diff_pathes(stage,from_commit, to_commit) do |path, content, mode|
            if mode == 40960 # symlink
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

      def archive_file_with_system_tar(stage, from_commit, to_commit)
        repo.dimg.tmp_path('archives', archive_file_name(from_commit, to_commit)).tap do |archive_path|
          relative_archive_file_path = File.join('archives_files', file_name(from_commit, to_commit))
          archive_diff_pathes(stage, from_commit, to_commit) do |path, content, mode|
            file_path = repo.dimg.tmp_path(relative_archive_file_path, path)

            if mode == 40960 # symlink
              FileUtils.symlink(content, file_path)
            else
              IO.write(file_path, content)
              FileUtils.chmod(mode, file_path)
            end
          end

          repo.dimg.dapp.shellout!("tar -C #{repo.dimg.tmp_path(relative_archive_file_path)} -cf #{archive_path} .")
        end
      end

      def archive_diff_pathes(stage, from_commit, to_commit)
        diff_patches(from_commit, to_commit).each do |patch|
          entry = patch.delta.new_file
          raise_if_submodule_patch!(patch)
          content = begin
            if to_commit == nil
              next unless (path = repo.path.dirname.join(entry[:path])).file?
              File.read(path)
            else
              repo.lookup_object(entry[:oid]).content
            end
          end

          yield slice_cwd(stage, entry[:path]), content, entry[:mode]
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

      def archive_file_name(from_commit, to_commit)
        file_name(from_commit, to_commit, ext: 'tar')
      end

      def patch_file(stage, from_commit, to_commit)
        File.open(repo.dimg.tmp_path('patches', patch_file_name(from_commit, to_commit)), File::RDWR | File::CREAT) do |f|
          diff_patches(from_commit, to_commit).each do |patch|
            raise_if_submodule_patch!(patch)
            f.write change_patch_new_file_path(stage, patch)
          end
        end
        repo.dimg.container_tmp_path('patches', patch_file_name(from_commit, to_commit))
      end

      def raise_if_submodule_patch!(patch) # FIXME
        entry = patch.delta.new_file
        if entry[:mode] == 57344 # submodule
          raise Error::Rugged, code: :submodule_not_supported, data: { path: repo.path.dirname.join(entry[:path]) }
        end
      end

      # rubocop:disable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity
      def change_patch_new_file_path(stage, patch)
        patch.to_s.lines.tap do |lines|
          modify_patch_line = proc do |line_number, path_char|
            action_part, path_part = lines[line_number].strip.split(' ', 2)
            if (path_with_cwd = path_part.partition("#{path_char}/").last).start_with?(cwd)
              native_path = case archive_type
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

      def diff_patches(from_commit, to_commit, paths: include_paths_or_cwd)
        (@diff_patches ||= {})[[from_commit, to_commit, paths]] ||= begin
          options = {}.tap do |opts|
            opts[:force_text] = true
            if dev_mode?
              opts[:include_untracked] = true
              opts[:recurse_untracked_dirs] = true
            end
          end
          repo.patches(from_commit, to_commit, paths: paths, exclude_paths: exclude_paths(true), **options)
        end
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

      def archive_stage_commits(stage)
        [nil, stage.layer_commit(self)]
      end

      def patch_stage_commits(stage)
        [stage.prev_g_a_stage.layer_commit(self), stage.layer_commit(self)]
      end

      def dev_patch_stage_commits(stage)
        [stage.prev_g_a_stage.layer_commit(self), nil]
      end

      def any_changes?(from_commit, to_commit)
        diff_patches(from_commit, to_commit).any?
      end

      def dev_mode?
        local? && repo.dimg.dev_mode?
      end

      def local?
        repo.is_a? GitRepo::Own
      end
    end
  end
end
