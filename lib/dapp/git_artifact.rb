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
      @cwd = (cwd.nil? || cwd.empty? || cwd == '/') ? '' : File.expand_path(File.join('/', cwd, '/'))[1..-1]
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

        if include_paths_or_cwd.empty? || include_paths_or_cwd.any? { |path| file_exist_in_repo?(stage.layer_commit(self), path) }
          commands << ["#{repo.dimg.project.git_bin} --git-dir=#{repo.container_path} archive #{stage.layer_commit(self)}:#{cwd} --prefix=/ #{include_paths.join(' ')}",
                       "#{sudo}#{repo.dimg.project.tar_bin} -x -C #{to} #{archive_command_excludes.join(' ')}"].join(' | ')
        end
      end
    end

    def apply_patch_command(stage)
      current_commit = stage.layer_commit(self)
      prev_commit = stage.prev_g_a_stage.layer_commit(self)

      if any_changes?(prev_commit, current_commit)
        [["#{repo.dimg.project.git_bin} --git-dir=#{repo.container_path} #{diff_command(prev_commit, current_commit)}",
          "#{sudo}#{repo.dimg.project.git_bin} apply --whitespace=nowarn --directory=#{to} #{patch_command_excludes.join(' ')} --unsafe-paths"].join(' | ')]
      else
        []
      end
    end

    def archive_command_excludes
      exclude_paths.map { |path| %(--exclude=#{File.join('/', path)}) }
    end

    def patch_command_excludes
      exclude_paths.map do |path|
        base = File.join(to, path)
        path =~ /[\*\?\[\]\{\}]/ ? %(--exclude=#{base} ) : %(--exclude=#{base} --exclude=#{File.join(base, '*')})
      end
    end

    def any_changes?(from, to = latest_commit)
      diff_patches(from, to).any?
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
      @latest_commit ||= (commit || repo.latest_commit(branch))
    end

    def paramshash
      Digest::SHA256.hexdigest [full_name, to, cwd, commit, branch, *include_paths, *exclude_paths, owner, group].map(&:to_s).join(':::')
    end

    def exclude_paths(with_cwd = false)
      base_paths(repo.exclude_paths + @exclude_paths, with_cwd)
    end

    def include_paths(with_cwd = false)
      base_paths(@include_paths, with_cwd)
    end

    def base_paths(paths, with_cwd = false)
      [paths].flatten.compact.map { |path| (with_cwd && cwd ? File.join(cwd, path) : path).gsub(%r{^\/*|\/*$}, '') }
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

    def diff_command(from, to, quiet: false)
      "diff --binary #{'--quiet' if quiet} #{from}..#{to} #{"--relative=#{cwd}" if cwd} -- #{include_paths(true).join(' ')}"
    end

    def include_paths_or_cwd
      case
      when !include_paths(true).empty? then include_paths(true)
      when !cwd.empty? then [cwd]
      else
        []
      end
    end

    def diff_patches(from, to)
      repo.diff(from, to, paths: include_paths_or_cwd).patches.select do |p|
        exclude_paths(true).any? { |path| !p.delta.new_file[:path].start_with?(path) }
      end
    end

    def file_exist_in_repo?(from, path)
      repo.file_exist_in_tree?(repo.lookup_commit(from).tree, path.split('/'))
    end
  end
end
