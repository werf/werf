module Dapp
  # Artifact from Git repo
  class GitArtifact
    include Dapp::CommonHelper

    MAX_PATCH_SIZE = 1024*1024

    # rubocop:disable Metrics/ParameterLists, Metrics/MethodLength
    def initialize(build, repo, where_to_add,
                   name: nil, branch: nil, commit: nil,
                   cwd: nil, paths: nil, owner: nil, group: nil,
                   interlayer_period: 7 * 24 * 3600, build_path: nil, flush_cache: false)
      @build = build
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

      @interlayer_period = interlayer_period

      @build_path = build_path || []

      @file_atomizer = build.builder.register_file_atomizer(build_path(filename('.file_atomizer')))
    end
    # rubocop:enable Metrics/ParameterLists, Metrics/MethodLength

    attr_reader :repo
    attr_reader :name
    attr_reader :where_to_add
    attr_reader :commit
    attr_reader :cwd
    attr_reader :owner
    attr_reader :group
    attr_reader :interlayer_period

    def archive_apply!(image, stage)
      layer_commit_write!(stage)
      layer_timestamp_write!(stage)

      credentials = [:owner, :group].map {|attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      image.build_cmd!(
          ["git --git-dir=#{repo.container_build_dir_path} archive",
           "--format tar.gz #{layer_commit(stage)}:#{cwd}",
           "-o #{container_archive_path} #{paths}"].join(' '),
          "mkdir -p #{where_to_add}",
          ["tar xf #{container_archive_path}", "-C #{where_to_add}", *credentials].join(' '),
          "rm -rf #{container_archive_path}"
      )
    end

    def layer_apply!(image, stage)
      return if stage.layer_actual?(self)

      layer_commit_write!(stage)
      layer_timestamp_write!(stage)

      apply_patch!(image, layer_commit(stage.prev_source_stage), layer_commit(stage))
    end

    def any_changes?(from, to = repo_latest_commit)
      !repo.git_bare("diff --quiet #{from}..#{to}#{" --relative=#{cwd}" if cwd} -- #{paths(true)}", returns: [0, 1]).status.success?
    end

    def patch_size_valid?(stage)
      patch_size(layer_commit(stage), layer_commit(stage.prev_source_stage)) < MAX_PATCH_SIZE
    end

    def layer_commit(stage)
      if layer_commit_file_path(stage).exist?
        layer_commit_file_path(stage).read.strip
      else
        layer_commit_write!(stage)
        repo_latest_commit
      end
    end

    def layer_timestamp(stage)
      if layer_timestamp_file_path(stage).exist?
        layer_timestamp_file_path(stage).read.strip.to_i
      else
        layer_timestamp_write!(stage)
        repo.commit_at(layer_commit(stage)).to_i
      end
    end

    def layer_commit_write!(stage)
      return if stage.name == :source_5

      file_atomizer.add_path(layer_commit_file_path(stage))
      layer_commit_file_path(stage).write(repo_latest_commit + "\n")
    end

    def layer_timestamp_write!(stage)
      return if stage.name == :source_5

      file_atomizer.add_path(layer_timestamp_file_path(stage))
      layer_timestamp_file_path(stage).write("#{repo.commit_at(layer_commit(stage)).to_i}\n")
    end

    private

    attr_reader :build
    attr_reader :file_atomizer

    def apply_patch!(image, from, to)
      image.build_cmd! "git --git-dir=#{repo.container_build_dir_path} diff #{from} #{to} | patch -l --directory=#{where_to_add}"
    end

    def patch_size(from, to)
      shellout("git --git-dir=#{repo.dir_path} diff #{from} #{to} | wc -c").stdout.strip.to_i
    end

    def container_archive_path
      container_build_path archive_filename
    end

    def layer_commit_filename(stage)
      layer_filename stage, '.commit'
    end

    def layer_timestamp_filename(stage)
      layer_filename stage, '.timestamp'
    end

    def layer_commit_file_path(stage)
      build_path layer_commit_filename(stage)
    end

    def layer_timestamp_file_path(stage)
      build_path layer_timestamp_filename(stage)
    end

    def build_path(*path)
      build.build_path(*@build_path, *path)
    end

    def container_build_path(*path)
      build.container_build_path(*@build_path, *path)
    end

    def archive_filename
      filename '.tar.gz'
    end

    def layer_filename(stage, ending)
      filename ".#{stage.name}.#{paramshash}.#{stage.git_artifact_signature}#{ending}"
    end

    def filename(ending)
      "#{repo.name}#{name ? "_#{name}" : nil}#{ending}"
    end

    def paramshash
      Digest::SHA256.hexdigest [cwd, paths, owner, group].map(&:to_s).join(':::')
    end

    def paths(with_cwd = false)
      [@paths].flatten.compact.map { |path| (with_cwd && cwd ? "#{cwd}/#{path}" : path).gsub(%r{^\/*|\/*$}, '') }.join(' ') if @paths
    end

    def repo_latest_commit
      commit
    end
  end
end
