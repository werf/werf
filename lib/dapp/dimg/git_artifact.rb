module Dapp
  module Dimg
    # Git repo artifact
    class GitArtifact
      attr_reader :repo
      attr_reader :name

      # rubocop:disable Metrics/ParameterLists
      def initialize(repo, to:, name: nil, branch: nil, commit: nil,
                     cwd: nil, include_paths: nil, exclude_paths: nil, owner: nil, group: nil,
                     stages_dependencies: {})
        @repo = repo
        @name = name

        @branch = branch || repo.dimg.dapp.cli_options[:git_artifact_branch] || repo.branch
        @commit = commit

        @to = to
        @cwd = begin
          if cwd.nil? || cwd.empty?
            ''
          else
            cwd = File.expand_path(File.join('/', cwd))[1..-1] # must be relative
            "#{cwd.chomp('/')}/"
          end
        end
        @include_paths = include_paths
        @exclude_paths = exclude_paths
        @owner = owner
        @group = group

        @stages_dependencies = stages_dependencies
      end
      # rubocop:enable Metrics/ParameterLists

      def apply_archive_command(stage)
        credentials = [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

        [].tap do |commands|
          commands << "#{repo.dimg.dapp.install_bin} #{credentials.join(' ')} -d #{to}"
          if any_changes?(nil, stage.layer_commit(self))
            commands << "#{sudo}#{repo.dimg.dapp.tar_bin} -xf #{archive_file(stage.layer_commit(self))} -C #{to}"
          end
        end
      end

      def apply_patch_command(stage)
        patch_command(stage.prev_g_a_stage.layer_commit(self), stage.layer_commit(self))
      end

      def apply_dev_patch_command(stage)
        patch_command(stage.prev_g_a_stage.layer_commit(self), nil)
      end

      def stage_dependencies_checksum(stage)
        return [] if (stage_dependencies = stages_dependencies[stage.name]).empty?

        paths = include_paths_or_cwd + base_paths(stage_dependencies, true)
        diff_patches(nil, latest_commit, paths: paths).map do |patch|
          delta_new_file = patch.delta.new_file
          Digest::SHA256.hexdigest [delta_new_file[:path], repo.lookup_object(delta_new_file[:oid]).content].join(':::')
        end
      end

      def patch_size(from, to)
        diff_patches(from, to).reduce(0) do |bytes, patch|
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
        Digest::SHA256.hexdigest diff_patches(stage.prev_g_a_stage.layer_commit(self), nil).map(&:to_s).join(':::')
      end

      def latest_commit
        @latest_commit ||= commit || repo.latest_commit(branch)
      end

      def paramshash
        Digest::SHA256.hexdigest [full_name, to, cwd, *include_paths, *exclude_paths, owner, group].map(&:to_s).join(':::')
      end

      def full_name
        "#{repo.name}#{name ? "_#{name}" : nil}"
      end

      def any_changes?(from, to = latest_commit)
        diff_patches(from, to).any?
      end

      protected

      attr_reader :to
      attr_reader :commit
      attr_reader :branch
      attr_reader :cwd
      attr_reader :owner
      attr_reader :group
      attr_reader :stages_dependencies

      def patch_command(prev_commit, current_commit)
        [].tap do |commands|
          if any_changes?(prev_commit, current_commit)
            commands << "#{sudo}#{repo.dimg.dapp.git_bin} apply --whitespace=nowarn --directory=#{to} --unsafe-paths " \
                        "#{patch_file(prev_commit, current_commit)}"
          end
        end
      end

      def sudo
        repo.dimg.dapp.sudo_command(owner: owner, group: group)
      end

      def archive_file(commit)
        create_file(repo.dimg.tmp_path('archives', archive_file_name(commit))) do |f|
          Gem::Package::TarWriter.new(f) do |tar|
            diff_patches(nil, commit).each do |patch|
              entry = patch.delta.new_file
              tar.add_file slice_cwd(entry[:path]), entry[:mode] do |tf|
                tf.write repo.lookup_object(entry[:oid]).content
              end
            end
          end
        end
        repo.dimg.container_tmp_path('archives', archive_file_name(commit))
      rescue Gem::Package::TooLongFileName => e
        raise Error::TarWriter, message: e.message
      end

      def slice_cwd(path)
        return path if cwd.empty?
        path
          .reverse
          .chomp(cwd.reverse)
          .reverse
      end

      def archive_file_name(commit)
        file_name(commit, 'tar')
      end

      def patch_file(from, to)
        create_file(repo.dimg.tmp_path('patches', patch_file_name(from, to))) do |f|
          diff_patches(from, to).each { |patch| f.write change_patch_new_file_path(patch) }
        end
        repo.dimg.container_tmp_path('patches', patch_file_name(from, to))
      end

      # rubocop:disable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity
      def change_patch_new_file_path(patch)
        patch.to_s.lines.tap do |lines|
          modify_patch_line = proc do |line_number, path_char|
            action_part, path_part = lines[line_number].split
            if (path_with_cwd = path_part.partition("#{path_char}/").last).start_with?(cwd)
              path_with_cwd.sub(cwd, '').tap do |native_path|
                expected_path = File.join(path_char, native_path)
                lines[line_number] = [action_part, expected_path].join(' ') + "\n"
              end
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

      def patch_file_name(from, to)
        file_name(from, to, 'patch')
      end

      def file_name(*args, ext)
        "#{[full_name, args].flatten.join('_')}.#{ext}"
      end

      def create_file(file_path, &blk)
        File.open(file_path, File::RDWR | File::CREAT, &blk)
      end

      def diff_patches(from, to, paths: include_paths_or_cwd)
        (@diff_patches ||= {})[[from, to, paths]] = repo.patches(from, to, paths: paths, exclude_paths: exclude_paths(true))
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
    end
  end
end
