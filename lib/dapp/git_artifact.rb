module Dapp
  # Artifact from Git repo
  class GitArtifact
    attr_reader :repo
    attr_reader :name

    # rubocop:disable Metrics/ParameterLists
    def initialize(repo, to:, name: nil, branch: nil, commit: nil,
                   cwd: nil, include_paths: nil, exclude_paths: nil, owner: nil, group: nil)
      @repo = repo
      @name = name

      @branch = branch || repo.dimg.project.cli_options[:git_artifact_branch] || repo.branch
      @commit = commit

      @to = to
      cwd = File.expand_path(File.join('/', cwd))[1..-1] unless cwd.nil? || cwd.empty? # must be relative!!!
      @cwd = cwd
      @include_paths = include_paths
      @exclude_paths = exclude_paths
      @owner = owner
      @group = group
    end
    # rubocop:enable Metrics/ParameterLists

    def apply_archive_command(stage)
      credentials = [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      [].tap do |commands|
        commands << "#{repo.dimg.project.install_bin} #{credentials.join(' ')} -d #{to}"
        commands << "#{sudo}#{repo.dimg.project.tar_bin} -xf #{archive_file(stage.layer_commit(self))} -C #{to}" if any_changes?(nil, stage.layer_commit(self))
      end
    end

    def apply_patch_command(stage)
      current_commit = stage.layer_commit(self)
      prev_commit = stage.prev_g_a_stage.layer_commit(self)

      [].tap do |commands|
        if any_changes?(prev_commit, current_commit)
          commands << "#{sudo}#{repo.dimg.project.git_bin} apply --whitespace=nowarn --directory=#{to} --unsafe-paths " \
                      "#{patch_file(prev_commit, current_commit)}"
        end
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

    def latest_commit
      @latest_commit ||= commit || repo.latest_commit(branch)
    end

    def paramshash
      Digest::SHA256.hexdigest [full_name, to, cwd, *include_paths, *exclude_paths, owner, group].map(&:to_s).join(':::')
    end

    def full_name
      "#{repo.name}#{name ? "_#{name}" : nil}"
    end

    protected

    attr_reader :to
    attr_reader :commit
    attr_reader :branch
    attr_reader :cwd
    attr_reader :owner
    attr_reader :group

    def sudo
      repo.dimg.project.sudo_command(owner: owner, group: group)
    end

    def archive_file(commit)
      create_file(repo.dimg.tmp_path('archives', archive_file_name(commit))) do |f|
        f.write begin
          StringIO.new.tap do |tar_stream|
            Gem::Package::TarWriter.new(tar_stream) do |tar|
              diff_patches(nil, commit).each do |patch|
                entry = patch.delta.new_file
                tar.add_file slice_cwd(entry[:path]), entry[:mode] do |tf|
                  tf.write repo.lookup_object(entry[:oid]).content
                end
              end
            end
          end.string
        end
      end
      repo.dimg.container_tmp_path('archives', archive_file_name(commit))
    rescue Gem::Package::TooLongFileName => e
      raise Error::TarWriter, message: e.message
    end

    def slice_cwd(path)
      return path if cwd.empty?
      path.gsub(/#{Regexp.escape(cwd)}\//, '')
    end

    def archive_file_name(commit)
      file_name(commit, 'tar')
    end

    def patch_file(from, to)
      create_file(repo.dimg.tmp_path('patches', patch_file_name(from, to))) do |f|
        diff_patches(from, to).each do |patch|
          f.write change_patch_new_file_path(patch.to_s)
        end
      end
      repo.dimg.container_tmp_path('patches', patch_file_name(from, to))
    end

    def change_patch_new_file_path(patch)
      patch.lines.tap do |lines|
        [0, 4].each do |ind|
          lines[ind] = begin
            line = lines[ind]
            new_path = line.split.last

            if !new_path.start_with?('/') && !cwd.empty?
              default_prefix = 'b'
              regex = /(?<=^#{default_prefix})\/#{Regexp.escape(cwd)}/
              new_path.gsub!(regex, '')
            end

            [line.split[0..-2], new_path].flatten.join(' ') + "\n"
          end
        end
      end.join
    end

    def patch_file_name(from, to)
      file_name(from, to, 'patch')
    end

    def file_name(*args, ext)
      "#{[full_name, args].flatten.join('_')}.#{ext}"
    end

    def create_file(file_path, &blk)
      File.open(file_path, File::RDWR|File::CREAT, &blk)
    end

    def any_changes?(from, to = latest_commit)
      diff_patches(from, to).any?
    end

    def diff_patches(from, to)
      (@diff_patches ||= {})[[from, to]] = repo.patches(from, to, paths: include_paths_or_cwd, exclude_paths: exclude_paths(true))
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
      base_paths(repo.dimg.project.system_files.concat(@exclude_paths), with_cwd)
    end

    def include_paths(with_cwd = false)
      base_paths(@include_paths, with_cwd)
    end

    def base_paths(paths, with_cwd = false)
      [paths].flatten.compact.map { |path| (with_cwd && cwd ? File.join(cwd, path) : path).gsub(%r{^\/*|\/*$}, '') }
    end
  end
end