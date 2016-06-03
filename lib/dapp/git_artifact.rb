module Dapp
  # Artifact from Git repo
  class GitArtifact
    include Dapp::CommonHelper

    # rubocop:disable Metrics/ParameterLists, Metrics/MethodLength
    def initialize(builder, repo, where_to_add,
                   name: nil, branch: nil, commit: nil,
                   cwd: nil, paths: nil, owner: nil, group: nil,
                   interlayer_period: 7 * 24 * 3600, build_path: nil, flush_cache: false)
      @builder = builder
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

      @atomizer = Atomizer.new builder, build_path(filename('.atomizer'))

      # check params hash
      lock do
        cleanup! unless !flush_cache && paramshash_path.exist? && paramshash_path.read.strip == paramshash
        paramshash_path.write paramshash + "\n"
      end
    end
    # rubocop:enable Metrics/ParameterLists, Metrics/MethodLength

    def signature
      hashsum [archive_commit, repo_latest_commit]
    end

    def build_path(*path)
      builder.build_path(*@build_path, *path)
    end

    def container_build_path(*path)
      builder.container_build_path(*@build_path, *path)
    end

    def cleanup!
      lock do
        FileUtils.rm_f [
                           paramshash_path,
                           archive_commit_file_path,
                           Dir.glob(layer_patch_path('*')),
                           Dir.glob(layer_commit_file_path('*')),
                           latest_commit_file_path
                       ].flatten
      end
    end

    def exist_in_step?(path, step)
      repo.exist_in_commit?(path, commit_by_step(step))
    end

    def prepare_step_commit
      archive_commit
    end

    def build_step_commit
      layer_commit(layers.last) || archive_commit
    end

    def setup_step_commit
      latest_commit || layer_commit(layers.last) || archive_commit
    end

    def commit_by_step(step)
      send :"#{step}_step_commit"
    end

    def any_changes?(from, to = repo_latest_commit)
      !repo.git_bare("diff --quiet #{from}..#{to}#{" --relative=#{cwd}" if cwd} -- #{paths(true)}", returns: [0, 1]).status.success?
    end

    attr_reader :repo
    attr_reader :name
    attr_reader :where_to_add
    attr_reader :commit
    attr_reader :cwd
    attr_reader :owner
    attr_reader :group
    attr_reader :interlayer_period

    def branch
      @branch || 'master'
    end

    protected

    attr_reader :builder
    attr_reader :atomizer

    def lock_with_repo(&blk)
      lock do
        repo.lock(&blk)
      end
    end

    def apply_layer_patches(image)
      latest_layer = nil
      layers.each do |layer|
        apply_layer_patch! image, layer
        latest_layer = layer
      end

      latest_layer
    end

    def paths(with_cwd = false)
      [@paths].flatten.compact.map { |path| (with_cwd && cwd ? "#{cwd}/#{path}" : path).gsub(%r{^\/*|\/*$}, '') }.join(' ') if @paths
    end

    def repo_latest_commit
      commit
    end

    def filename(ending)
      "#{repo.name}#{name ? "_#{name}" : nil}#{ending}"
    end

    def paramshash_filename
      filename '.paramshash'
    end

    def paramshash_path
      build_path paramshash_filename
    end

    def paramshash
      Digest::SHA256.hexdigest [cwd, paths, owner, group].map(&:to_s).join(':::')
    end

    def archive_filename
      filename '.tar.gz'
    end

    def container_archive_path
      container_build_path archive_filename
    end

    def archive_commit_timestamp_filename
      filename '.timestamp'
    end

    def archive_commit_timestamp_path
      build_path archive_commit_timestamp_filename
    end

    def container_archive_commit_timestamp_path
      container_build_path archive_commit_timestamp_filename
    end

    def archive_commit_file_filename
      filename '.commit'
    end

    def archive_commit_file_path
      build_path archive_commit_file_filename
    end

    def container_archive_commit_file_path
      container_build_path archive_commit_file_filename
    end

    def archive_commit
      if archive_commit_file_path.exist?
        archive_commit_file_path.read.strip
      else
        repo_latest_commit
      end
    end

    def archive_commit_file_exist?
      archive_commit_file_path.exist?
    end

    def apply_archive!(image)
      return if archive_commit_file_exist?

      atomizer << archive_commit_file_path
      atomizer << archive_commit_timestamp_path

      archive_commit_file_path.write archive_commit + "\n"
      archive_commit_timestamp_path.write repo.commit_at(archive_commit) + "\n"

      credentials = [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      image.build_cmd!(
        "git --git-dir=#{repo.container_build_dir_path} archive --format tar.gz #{archive_commit}:#{cwd} -o #{container_archive_path} #{paths}",
        "mkdir -p #{where_to_add}",
        ["tar xf #{container_archive_path} ", "-C #{where_to_add} ", "--strip-components=1", *credentials].join,
        "rm -rf #{container_archive_path}"
      )
    end

    def sudo_format_user(user)
      user.to_i.to_s == user ? "\\\##{user}" : user
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

    def apply_patch!(image, from, to)
      image.build_cmd! "git diff #{from} #{to} | git apply --whitespace=nowarn --directory=#{where_to_add}"
    end

    def layer_filename(stage, ending)
      filename "_layer_#{stage.is_a?(Fixnum) ? format('%04d', stage) : stage}#{ending}"
    end

    def layer_commit_file_path(stage)
      build_path layer_filename(stage, '.commit')
    end

    def layer_actual?(stage)
      layer_commit(stage) == archive_commit || !any_changes?(archive_commit, layer_commit(stage))
    end

    def layer_commit(stage)
      if layer_commit_file_path(stage).exist?
        layer_commit_file_path(stage).read.strip
      else
        repo_latest_commit
      end
    end

    def source_1_commit
      layer_commit(:source_1)
    end

    def layers
      Dir.glob(layer_commit_file_path('*')).map { |path| Integer(path.gsub(/.*_(\d+)\.commit$/, '\\1')) }.sort
    end

    def apply_layer_patch!(image, stage)
      return if layer_actual?(stage)
      # apply_patch! image, layer_patch_filename(stage)
      # TODO
    end

    def latest_commit_file_path
      build_path filename '_latest.commit'
    end

    def latest_commit
      latest_commit_file_path.read.strip if latest_commit_file_path.exist?
    end

    def apply_latest_patch!(image)
      # apply_patch! image, latest_patch_filename
      # TODO
    end

    def remove_latest!
      FileUtils.rm_f [latest_commit_file_path]
    end

    def lock(**kwargs, &blk)
      builder.filelock(build_path(filename('.lock')),
                       error_message: "Git artifact commit:#{commit}" +
                           "#{name ? " #{name}" : nil} #{repo.name}" +
                           " (#{repo.dir_path}) in use! Try again later.",
                       **kwargs, &blk)
    end
  end
end
