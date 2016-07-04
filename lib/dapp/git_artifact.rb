module Dapp
  # Artifact from Git repo
  class GitArtifact
    include Dapp::CommonHelper

    attr_reader :repo
    attr_reader :name

    # rubocop:disable Metrics/ParameterLists, Metrics/MethodLength
    def initialize(repo, where_to_add,
                   name: nil, branch: nil, commit: nil,
                   cwd: nil, paths: nil, owner: nil, group: nil)
      @repo = repo
      @name = name

      @where_to_add = where_to_add

      @branch = branch
      @commit = commit

      @cwd = cwd
      @paths = paths
      @owner = owner
      @group = group
    end
    # rubocop:enable Metrics/ParameterLists, Metrics/MethodLength

    def archive_apply_command(stage)
      credentials = [:owner, :group].map {|attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      ["install #{credentials.join(' ')} -d #{where_to_add}",
       ["git --git-dir=#{repo.container_build_dir_path} archive #{stage.layer_commit(self)}:#{cwd} #{paths}",
       "#{sudo}tar -x -C #{where_to_add}"].join(' | ')]
    end

    def apply_patch_command(stage)
      current_commit = stage.layer_commit(self)
      prev_commit = stage.prev_source_stage.layer_commit(self)

      if prev_commit != current_commit or any_changes?(prev_commit, current_commit)
        [["git --git-dir=#{repo.container_build_dir_path} #{diff_command(prev_commit, current_commit)}",
         "#{sudo}git apply --whitespace=nowarn --directory=#{where_to_add} " \
         "$(if [ \"$(git --version)\" != \"git version 1.9.1\" ]; then echo \"--unsafe-paths\"; fi)"].join(' | ')] # FIXME
      else
        []
      end
    end

    def any_changes?(from, to=latest_commit)
      !repo.git_bare(diff_command(from, to, quiet: true), returns: [0, 1]).status.success?
    end

    def patch_size(from, to)
      repo.git_bare("#{diff_command(from, to)} | wc -c").stdout.strip.to_i
    end

    def latest_commit
      @latest_commit ||= commit || repo.latest_commit(branch)
    end

    def paramshash
      Digest::SHA256.hexdigest [cwd, paths, owner, group].map(&:to_s).join(':::')
    end

    def paths(with_cwd = false)
      [@paths].flatten.compact.map { |path| (with_cwd && cwd ? "#{cwd}/#{path}" : path).gsub(%r{^\/*|\/*$}, '') }.join(' ') if @paths
    end

    def filename(ending)
      "#{repo.name}#{name ? "_#{name}" : nil}#{ending}"
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
        sudo = 'sudo '
        sudo += "-u #{sudo_format_user(owner)} " if owner
        sudo += "-g #{sudo_format_user(group)} " if group
      end

      sudo
    end

    def diff_command(from, to, quiet: false)
      "diff #{'--quiet' if quiet } #{from}..#{to} #{"--relative=#{cwd}" if cwd} -- #{paths(true)}"
    end
  end
end
