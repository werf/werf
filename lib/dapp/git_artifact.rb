module Dapp
  # Artifact from Git repo
  class GitArtifact
    include Dapp::CommonHelper

    attr_reader :repo
    attr_reader :name
    attr_reader :where_to_add
    attr_reader :commit
    attr_reader :cwd
    attr_reader :owner
    attr_reader :group

    # rubocop:disable Metrics/ParameterLists, Metrics/MethodLength
    def initialize(repo, where_to_add,
                   name: nil, branch: nil, commit: nil,
                   cwd: nil, paths: nil, owner: nil, group: nil)
      @repo = repo
      @name = name

      @where_to_add = where_to_add

      @branch = branch
      @commit = commit || begin
        @branch ? repo.latest_commit(branch) : repo.latest_commit('HEAD')
      end

      @cwd = cwd
      @paths = paths
      @owner = owner
      @group = group
    end
    # rubocop:enable Metrics/ParameterLists, Metrics/MethodLength

    def archive_apply_command(stage)
      credentials = [:owner, :group].map {|attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      [["git --git-dir=#{repo.container_build_dir_path} archive",
       "--format tar.gz #{stage.layer_commit(self)}:#{cwd}",
       "-o #{stage.container_archive_path(self)} #{paths}"].join(' '),
       "mkdir -p #{where_to_add}",
       ["tar xf #{stage.container_archive_path(self)}", "-C #{where_to_add}", *credentials].join(' '),
       "rm -rf #{stage.container_archive_path(self)}"]
    end

    def apply_patch_command(stage)
      current_commit = stage.layer_commit(self)
      prev_commit = stage.prev_source_stage.layer_commit(self)

      if prev_commit != current_commit or any_changes?(prev_commit, current_commit)
        ["git --git-dir=#{repo.container_build_dir_path} diff #{prev_commit} #{current_commit} | " \
         "git apply --whitespace=nowarn --directory=#{where_to_add} " \
         "$(if [ \"$(git --version)\" != \"git version 1.9.1\" ]; then echo \"--unsafe-paths\"; fi)"] # FIXME
      else
        []
      end
    end

    def any_changes?(from, to=repo_latest_commit)
      !repo.git_bare("diff --quiet #{from}..#{to}#{" --relative=#{cwd}" if cwd} -- #{paths(true)}", returns: [0, 1]).status.success?
    end

    def patch_size(from, to)
      shellout!("git --git-dir=#{repo.dir_path} diff #{from} #{to} | wc -c").stdout.strip.to_i
    end

    def repo_latest_commit
      commit
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
  end
end
