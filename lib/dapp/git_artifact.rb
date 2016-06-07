module Dapp
  # Artifact from Git repo
  class GitArtifact
    include Dapp::CommonHelper

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
      build.build_path(*@build_path, *path)
    end

    def container_build_path(*path)
      build.container_build_path(*@build_path, *path)
    end

    def cleanup!
      lock do
        FileUtils.rm_f [
          paramshash_path,
          archive_commit_file_path,
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

    def layer_filename(stage, ending)
      filename "#{stage}.#{build.stages[stage].signature}#{ending}"
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

    def layer_commit(stage)
      if layer_commit_filename(stage).exist?
        layer_commit_filename(stage).read.strip 
      else
        repo_latest_commit
      end
    end

    def layer_commit_write!(stage)
      file_atomizer.add_file(layer_commit_file_path(stage))
      layer_commit_file_path(stage).write(layer_commit(stage) + "\n")
    end

    def layer_timestamp(stage)
      if layer_timestamp_filename(stage).exist?
        layer_timestamp_filename(stage).read.strip.to_i
      else
        repo.commit_at(layer_commit(stage))
      end
    end

    def layer_timestamp_write!(stage)
      return unless layer_timestamp_file_path(stage).zero?

      file_atomizer.add_file(layer_timestamp_file_path(stage))
      layer_timestamp_file_path(stage).write(layer_timestamp(stage) + "\n")
    end

    def layer_actual?(stage)
      layer_commit(stage) != layer_commit(layer_prev_stage(stage)) and
        any_changes?(archive_commit, layer_commit(stage))
    end

    def layer_exist?(stage)
      layer_commit_file_path(stage).exist?
    end

    def layer_prev_stage(stage)
      s = stage
      while (prev_stage = build.stages[s].prev)
        return prev_stage.name if layer_commit(prev_stage.name)
        s = prev_stage
      end
      nil
    end

    def layer_patch(stage)
      ''
    end

    def layer_apply!(image, stage)
      return if layer_actual?(stage)

      layer_commit_write!(stage)
      layer_timestamp_write!(stage)
      apply_patch!(image, layer_commit(layer_prev_stage(stage)), layer_commit(stage))
    end

    %i(source_1_archive source_1 source_2 source_3 source_4 source_5).each do |stage|
      define_method("#{stage}_actual?") {layer_actual?(stage)}
      define_method("#{stage}_patch") {layer_patch(stage)}
      define_method("#{stage}_commit") {layer_commit(stage)}
      define_method("#{stage}_timestamp") {layer_timestamp(stage)}
      define_method("#{stage}_apply!") {|image| layer_apply!(image, stage)}
    end

    def source_1_archive_apply!(image)
      return if layer_commit_file_path(:source_1_archive).exist?

      layer_commit_write!(:source_1_archive)
      layer_timestamp_write!(:source_1_archive)

      credentials = [:owner, :group].map {|attr|
        "--#{attr}=#{send(attr)}" unless send(attr).nil?
      }.compact

      image.build_cmd!(
        ["git --git-dir=#{repo.container_build_dir_path} archive",
         "--format tar.gz #{archive_commit}:#{cwd}",
         "-o #{container_archive_path} #{paths}"].join(' '),
        "mkdir -p #{where_to_add}",
        ["tar xf #{container_archive_path}", "-C #{where_to_add}",
         '--strip-components=1', *credentials].join(' '),
        "rm -rf #{container_archive_path}",
      )
    end

    def source_4_actual? # FIXME: skipped stage
      true
    end

    def source_4_apply!(image) # FIXME: skipped stage
    end

    protected

    attr_reader :build
    attr_reader :file_atomizer

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

    def archive_timestamp_filename
      filename '.timestamp'
    end

    def archive_timestamp_path
      build_path archive_timestamp_filename
    end

    def archive_timestamp
      value = nil
      value = archive_timestamp_path.read.strip.to_i if archive_timestamp_path.exist?
      value
    end

    def container_archive_timestamp_path
      container_build_path archive_timestamp_filename
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
      image.build_cmd! "git --git-dir=#{repo.container_build_dir_path} diff #{from} #{to} | git --git-dir=#{repo.container_build_dir_path} apply --whitespace=nowarn --directory=#{where_to_add}"
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
      build.filelock(build_path(filename('.lock')),
                     error_message: "Git artifact commit:#{commit}" +
                         "#{name ? " #{name}" : nil} #{repo.name}" +
                         " (#{repo.dir_path}) in use! Try again later.",
                     **kwargs, &blk)
    end
  end
end
