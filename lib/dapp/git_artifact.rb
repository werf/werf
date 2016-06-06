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
      filename "#{stage}.#{builder.stages[stage].signature}#{ending}"
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

    def layer_timestamp(stage)
      if layer_timestamp_filename(stage).exist?
        layer_timestamp_filename(stage).read.strip.to_i
      else
        repo.commit_at(layer_commit(stage))
      end
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
      while prev_stage = builder.stages[s].prev
        return prev_stage.name if layer_commit(prev_stage.name)
        s = prev_stage
      end
    end


    def source_1_actual?
      layer_actual?(:source_1)
    end

    def source_2_actual?
      layer_actual?(:source_2)
    end

    def source_3_actual?
      layer_actual?(:source_3)
    end

    def source_4_actual?
      layer_actual?(:source_4)
    end

    def source_5_actual?
      if source_4_commit
        source_5_commit == source_4_commit || !any_changes?(source_4_commit, source_5_commit)
      else
        source_5_commit == source_3_commit || !any_changes?(source_3_commit, source_5_commit)
      end
    end


    def source_4_patch
      ''
    end

    def source_5_patch
      ''
    end


    def apply_source_1_archive!(image)
      return if archive_commit_file_exist?

      atomizer << archive_commit_file_path
      atomizer << archive_timestamp_path

      archive_commit_file_path.write archive_commit + "\n"
      archive_timestamp_path.write repo.commit_at(archive_commit).to_s + "\n"

      credentials = [:owner, :group].map { |attr| "--#{attr}=#{send(attr)}" unless send(attr).nil? }.compact

      image.build_cmd!(
          "git --git-dir=#{repo.container_build_dir_path} archive --format tar.gz #{archive_commit}:#{cwd} -o #{container_archive_path} #{paths}",
          "mkdir -p #{where_to_add}",
          ["tar xf #{container_archive_path} ", "-C #{where_to_add} ", "--strip-components=1", *credentials].join,
          "rm -rf #{container_archive_path}"
      )
    end

    def apply_source_1!(image)
      if layer_timestamp(:source_1).to_i < archive_timestamp.to_i
        delete_file(layer_commit_file_path(:source_1))
        delete_file(layer_timestamp_file_path(:source_1))
      end

      atomizer << layer_commit_file_path(:source_1)
      atomizer << layer_timestamp_file_path(:source_1)

      layer_commit_file_path(:source_1).write repo_latest_commit + "\n"
      layer_timestamp_file_path(:source_1).write repo.commit_at(layer_commit(:source_1)).to_s + "\n" if layer_timestamp_file_path(:source_1).zero?
      apply_patch!(image, archive_commit, layer_commit(:source_1)) if layer_actual?(:source_1)
    end

    def apply_source_2!(image)
      if layer_timestamp(:source_2).to_i < layer_timestamp(:source_1).to_i
        delete_file(layer_commit_file_path(:source_2))
        delete_file(layer_timestamp_file_path(:source_2))
      end

      atomizer << layer_commit_file_path(:source_2)
      atomizer << layer_timestamp_file_path(:source_2)

      layer_commit_file_path(:source_2).write repo_latest_commit + "\n"
      layer_timestamp_file_path(:source_2).write repo.commit_at(layer_commit(:source_2)).to_s + "\n" if layer_timestamp_file_path(:source_2).zero?
      apply_patch!(image, layer_commit(:source_1), layer_commit(:source_2)) if layer_actual?(:source_2)
    end

    def apply_source_3!(image)
      if layer_timestamp(:source_3).to_i < layer_timestamp(:source_2).to_i
        delete_file(layer_commit_file_path(:source_3))
        delete_file(layer_timestamp_file_path(:source_3))
      end

      atomizer << layer_commit_file_path(:source_3)
      atomizer << layer_timestamp_file_path(:source_3)

      layer_commit_file_path(:source_3).write repo_latest_commit + "\n"
      layer_timestamp_file_path(:source_3).write repo.commit_at(layer_commit(:source_3)).to_s + "\n" if layer_timestamp_file_path(:source_3).zero?
      apply_patch!(image, layer_commit(:source_2), layer_commit(:source_3)) if layer_actual?(:source_3)
    end

    def apply_source_4!(image)
      if source_4_timestamp.to_i < source_3_timestamp.to_i
        layer_commit_file_path(:source_4).delete
        layer_timestamp_file_path(:source_4).delete
      end

      return if layer_actual?(:source_4)

      atomizer << layer_commit_file_path(:source_4)
      atomizer << layer_timestamp_file_path(:source_4)

      layer_commit_file_path(:source_4).write source_4_commit + "\n"
      layer_timestamp_file_path(:source_4).write repo.commit_at(source_4_commit) + "\n" if layer_timestamp_file_path(:source_4).zero?

      apply_patch!(image, source_3_commit, source_4_commit)
    end

    def apply_source_5!(image)
      if (source_4_commit and source_5_timestamp.to_i < source_4_timestamp.to_i) or
         (source_5_timestamp.to_i < source_3_timestamp.to_i)
        layer_commit_file_path(:source_5).delete
        layer_timestamp_file_path(:source_5).delete
      end

      return if layer_actual?(:source_5)

      atomizer << layer_commit_file_path(:source_5)
      atomizer << layer_timestamp_file_path(:source_5)

      layer_commit_file_path(:source_5).write source_5_commit + "\n"
      layer_timestamp_file_path(:source_5).write repo.commit_at(source_5_commit) + "\n" if layer_timestamp_file_path(:source_5).zero?

      if source_4_commit
        apply_patch!(image, source_4_commit, source_5_commit)
      else
        apply_patch!(image, source_3_commit, source_5_commit)
      end
    end

    def source_5_prev_stage
      %i(source_4 source_3 source_2 source_1 source_1_archive).each do |stage|
        return stage if layer_commit(stage)
      end
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

    def layer_filename(stage, ending)
      filename "_layer_#{stage.is_a?(Fixnum) ? format('%04d', stage) : stage}#{ending}"
    end

    def layer_commit_file_path(stage)
      build_path layer_filename(stage, '.commit')
    end

    def layer_timestamp_file_path(stage)
      build_path layer_filename(stage, '.timestamp')
    end

    def layer_commit(stage)
      if layer_commit_file_path(stage).exist? and layer_timestamp_file_path(stage).exist?
        layer_commit_file_path(stage).read.strip
      else
        repo_latest_commit
      end
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

    def layer_timestamp(stage)
      value = nil
      value = layer_timestamp_file_path(stage).read.strip.to_i if layer_timestamp_file_path(stage).exist?
      value
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
