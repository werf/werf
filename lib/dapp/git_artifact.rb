module Dapp
  # Artifact from Git repo
  class GitArtifact
    attr_reader :repo
    attr_reader :name

    # rubocop:disable Metrics/ParameterLists
    def initialize(repo, where_to_add:, name: nil, branch: nil, commit: nil,
                   cwd: nil, paths: nil, exclude_paths: nil, owner: nil, group: nil)
      @repo = repo
      @name = name

      @where_to_add = where_to_add

      @branch = branch || repo.application.project.cli_options[:git_artifact_branch] || repo.branch
      @commit = commit

      cwd = File.expand_path(File.join('/', cwd))[1..-1] unless cwd.nil? || cwd.empty?
      @cwd = cwd
      @paths = paths
      @exclude_paths = exclude_paths
      @owner = owner
      @group = group
    end
    # rubocop:enable Metrics/ParameterLists

    def apply_archive_command(stage)
      credentials = [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      ["install #{credentials.join(' ')} -d #{where_to_add}",
       ["#{repo.application.git_path} --git-dir=#{repo.container_path} archive #{stage.layer_commit(self)}:#{cwd} #{paths.join(' ')}",
        "#{sudo}tar -x -C #{where_to_add} #{archive_command_excludes.join(' ')}"].join(' | ')]
    end

    def apply_patch_command(stage)
      current_commit = stage.layer_commit(self)
      prev_commit = stage.prev_g_a_stage.layer_commit(self)

      if prev_commit != current_commit || any_changes?(prev_commit, current_commit)
        [["#{repo.application.git_path} --git-dir=#{repo.container_path} #{diff_command(prev_commit, current_commit)}",
          "#{sudo}#{repo.application.git_path} apply --whitespace=nowarn --directory=#{where_to_add} #{patch_command_excludes.join(' ')} --unsafe-paths"].join(' | ')]
      else
        []
      end
    end

    def archive_command_excludes
      exclude_paths.map { |path| %(--exclude=#{path}) }
    end

    def patch_command_excludes
      exclude_paths.map do |path|
        base = File.join(where_to_add, path)
        (path =~ /[\*\?\[\]\{\}]/) ? %(--exclude=#{base} ) : %(--exclude=#{base} --exclude=#{File.join(base, '*')})
      end
    end

    def any_changes?(from, to = latest_commit)
      !repo.git_bare(diff_command(from, to, quiet: true), returns: [0, 1]).status.success?
    end

    def patch_size(from, to)
      repo.git_bare("#{diff_command(from, to)} | wc -c").stdout.strip.to_i
    end

    def latest_commit
      @latest_commit ||= commit || repo.latest_commit(branch)
    end

    def paramshash
      Digest::SHA256.hexdigest [where_to_add, cwd, *paths, *exclude_paths, owner, group].map(&:to_s).join(':::')
    end

    def exclude_paths(with_cwd = false)
      base_paths(@exclude_paths, with_cwd = with_cwd)
    end

    def paths(with_cwd = false)
      base_paths(@paths, with_cwd = with_cwd)
    end

    def base_paths(paths, with_cwd = false)
      [paths].flatten.compact.map { |path| (with_cwd && cwd ? File.join(cwd, path) : path).gsub(%r{^\/*|\/*$}, '') }
    end

    def full_name
      "#{repo.name}#{name ? "_#{name}" : nil}"
    end

    protected

    attr_reader :where_to_add
    attr_reader :commit
    attr_reader :branch
    attr_reader :cwd
    attr_reader :owner
    attr_reader :group

    def sudo_format_user(user)
      user.to_s.to_i.to_s == user ? "\\\##{user}" : user
    end

    def sudo
      sudo = ''

      if owner || group
        sudo = "#{repo.application.sudo_path} -E "
        sudo += "-u #{sudo_format_user(owner)} " if owner
        sudo += "-g #{sudo_format_user(group)} " if group
      end

      sudo
    end

    def diff_command(from, to, quiet: false)
      "diff --binary #{'--quiet' if quiet} #{from}..#{to} #{"--relative=#{cwd}" if cwd} -- #{paths(true).join(' ')}"
    end
  end
end
