module Dapp
  # Artifact from Git repo
  class GitArtifact
    include Dapp::CommonHelper

    # rubocop:disable Metrics/ParameterLists, Metrics/MethodLength
    def initialize(builder, repo, where_to_add, name: nil, branch: 'master', cwd: nil, paths: nil, owner: nil, group: nil,
                   interlayer_period: 7 * 24 * 3600, build_path: nil, flush_cache: false)
      @builder = builder
      @repo = repo
      @name = name

      @where_to_add = where_to_add
      @branch = branch
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

    def add_multilayer!(image)
      lock_with_repo do
        # create and add archive
        create_and_add_archive(image)
        return if archive_commit == repo_latest_commit

        # add layer patches
        latest_layer = add_layer_patches(image)
        if latest_layer
          latest_layer_commit = layer_commit(latest_layer)
          return if latest_layer_commit == repo_latest_commit
        end

        # empty changes
        unless any_changes?(latest_layer_commit || archive_commit)
          remove_latest!
          return
        end

        # create and add last patch
        create_and_add_last_patch(image, latest_layer, latest_layer_commit)
      end
    end

    def cleanup!
      lock do
        FileUtils.rm_f [
          paramshash_path,
          archive_path,
          archive_commitfile_path,
          Dir.glob(layer_patch_path('*')),
          Dir.glob(layer_commitfile_path('*')),
          latest_patch_path,
          latest_commitfile_path
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

    def any_changes?(from)
      !repo.git_bare("diff --quiet #{from}..#{repo_latest_commit}#{" --relative=#{cwd}" if cwd} -- #{paths(true)}", returns: [0, 1]).status.success?
    end

    attr_reader :repo
    attr_reader :name
    attr_reader :where_to_add
    attr_reader :branch
    attr_reader :cwd
    attr_reader :owner
    attr_reader :group
    attr_reader :interlayer_period

    protected

    attr_reader :builder
    attr_reader :atomizer

    def lock_with_repo(&blk)
      lock do
        repo.lock(&blk)
      end
    end

    def create_and_add_archive(image)
      create_archive! unless archive_exist?
      add_archive(image)
    end

    def add_layer_patches(image)
      latest_layer = nil
      layers.each do |layer|
        add_layer_patch image, layer
        latest_layer = layer
      end

      latest_layer
    end

    def create_and_add_last_patch_as_layer_patch(image, latest_layer, latest_layer_commit)
      remove_latest!
      layer = latest_layer.to_i + 1
      create_layer_patch!(latest_layer_commit || archive_commit, layer)
      add_layer_patch image, layer
    end

    def create_and_add_last_patch_as_latest_patch(image, _latest_layer, latest_layer_commit)
      if latest_commit != repo_latest_commit
        create_latest_patch!(latest_layer_commit || archive_commit)
      end
      add_latest_patch(image)
    end

    def create_and_add_last_patch(image, latest_layer, latest_layer_commit)
      if (Time.now - repo.commit_at(latest_layer_commit || archive_commit)) > interlayer_period
        create_and_add_last_patch_as_layer_patch(image, latest_layer, latest_layer_commit)
      else
        create_and_add_last_patch_as_latest_patch(image, latest_layer, latest_layer_commit)
      end
    end

    def paths(with_cwd = false)
      [@paths].flatten.compact.map { |path| (with_cwd && cwd ? "#{cwd}/#{path}" : path).gsub(%r{^\/*|\/*$}, '') }.join(' ') if @paths
    end

    def repo_latest_commit
      repo.latest_commit(branch)
    end

    def filename(ending)
      "#{repo.name}#{name ? "_#{name}" : nil}.#{branch}#{ending}"
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

    def archive_path
      build_path archive_filename
    end

    def container_archive_path
      container_build_path archive_filename
    end

    def archive_commitfile_filename
      filename '.commit'
    end

    def archive_commitfile_path
      build_path archive_commitfile_filename
    end

    def container_archive_commitfile_path
      container_build_path archive_commitfile_filename
    end

    def archive_commit
      if archive_commitfile_path.exist?
        archive_commitfile_path.read.strip
      else
        repo_latest_commit
      end
    end

    def create_arhive_with_owner_substitution!
      Dir.mktmpdir('dapp_change_archive_owner') do |tmpdir_path|
        atomizer << tmpdir_path
        repo.git_bare "archive #{repo_latest_commit}:#{cwd} #{paths} | /bin/tar --extract --directory #{tmpdir_path}"
        builder.shellout("/usr/bin/find #{tmpdir_path} -maxdepth 1 -mindepth 1 -printf '%P\\n' | /bin/tar -czf #{archive_path} -C #{tmpdir_path}" \
                         " -T - --owner=#{owner || 'root'} --group=#{group || 'root'}")
      end
    end

    def create_simple_archive!
      repo.git_bare "archive --format tar.gz #{repo_latest_commit}:#{cwd} -o #{archive_path} #{paths}"
    end

    def create_archive!
      atomizer << archive_path
      atomizer << archive_commitfile_path

      if owner || group
        create_arhive_with_owner_substitution!
      else
        create_simple_archive!
      end

      archive_commitfile_path.write archive_commit + "\n"
    end

    def archive_exist?
      archive_commitfile_path.exist?
    end

    def add_archive(image)
      image.build_cmd!("mkdir -p #{where_to_add}",
                       ["tar xf #{container_archive_path} ",
                        "-C #{where_to_add} ",
                        "--strip-components=1"].join)
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

    def add_patch(image, filename)
      image.build_cmd! ["zcat #{container_build_path(filename)} | ",
                        "#{sudo}git apply --whitespace=nowarn --directory=#{where_to_add}"].join
    end

    def create_patch!(from, filename, commitfile_path)
      atomizer << build_path(filename)
      atomizer << commitfile_path

      repo.git_bare "diff --binary #{from}..#{repo_latest_commit}" +
                    "#{" --relative=#{cwd}" if cwd} -- #{paths(true)} " +
                    "| gzip > #{build_path filename}"

      commitfile_path.write repo_latest_commit + "\n"
    end

    def layer_filename(layer, ending)
      filename "_layer_#{layer.is_a?(Fixnum) ? format('%04d', layer) : layer}#{ending}"
    end

    def layer_patch_filename(layer)
      layer_filename(layer, '.patch.gz')
    end

    def layer_patch_path(layer)
      build_path layer_patch_filename(layer)
    end

    def layer_commitfile_path(layer)
      build_path layer_filename(layer, '.commit')
    end

    def layer_commit(layer)
      layer_commitfile_path(layer).read.strip if layer_commitfile_path(layer).exist?
    end

    def layers
      Dir.glob(layer_commitfile_path('*')).map { |path| Integer(path.gsub(/.*_(\d+)\.commit$/, '\\1')) }.sort
    end

    def create_layer_patch!(from, layer)
      create_patch! from, layer_patch_filename(layer), layer_commitfile_path(layer)
    end

    def add_layer_patch(image, layer)
      add_patch image, layer_patch_filename(layer)
    end

    def latest_patch_filename
      filename '_latest.patch.gz'
    end

    def latest_patch_path
      build_path latest_patch_filename
    end

    def latest_commitfile_path
      build_path filename '_latest.commit'
    end

    def latest_commit
      latest_commitfile_path.read.strip if latest_commitfile_path.exist?
    end

    def create_latest_patch!(from)
      create_patch! from, latest_patch_filename, latest_commitfile_path
    end

    def add_latest_patch(image)
      add_patch image, latest_patch_filename
    end

    def remove_latest!
      FileUtils.rm_f [latest_patch_path, latest_commitfile_path]
    end

    def lock(**kwargs, &block)
      builder.filelock(build_path(filename('.lock')), error_message: "Branch #{branch} of artifact #{name ? " #{name}" : nil} #{repo.name}" \
                       " (#{repo.dir_path}) in use! Try again later.", **kwargs, &block)
    end
  end
end
