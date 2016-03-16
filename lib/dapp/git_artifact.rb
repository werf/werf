module Dapp
  # Artifact from Git repo
  class GitArtifact
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
        cleanup! unless !flush_cache && File.exist?(paramshash_path) && File.read(paramshash_path) == paramshash
        File.write paramshash_path, paramshash
      end
    end
    # rubocop:enable Metrics/ParameterLists, Metrics/MethodLength

    def build_path(*paths)
      builder.build_path(*@build_path, *paths)
    end

    def add_multilayer!
      lock_with_repo do
        # create and add archive
        create_and_add_archive
        return if archive_commit == repo_latest_commit

        # add layer patches
        latest_layer = add_layer_patches
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
        create_and_add_last_patch(latest_layer, latest_layer_commit)
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
      !repo.git_bare("diff --quiet #{from}..#{repo_latest_commit}#{" --relative=#{cwd}" if cwd} #{paths(true)}", returns: [0, 1]).status.success?
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

    def create_and_add_archive
      create_archive! unless archive_exist?
      add_archive
    end

    def add_layer_patches
      latest_layer = nil
      layers.each do |layer|
        add_layer_patch layer
        latest_layer = layer
      end

      latest_layer
    end

    def create_and_add_last_patch_as_layer_patch(latest_layer, latest_layer_commit)
      remove_latest!
      layer = latest_layer.to_i + 1
      create_layer_patch!(latest_layer_commit || archive_commit, layer)
      add_layer_patch layer
    end

    def create_and_add_last_patch_as_latest_patch(_latest_layer, latest_layer_commit)
      if latest_commit != repo_latest_commit
        create_latest_patch!(latest_layer_commit || archive_commit)
      end
      add_latest_patch
    end

    def create_and_add_last_patch(latest_layer, latest_layer_commit)
      if (Time.now - repo.commit_at(latest_layer_commit || archive_commit)) > interlayer_period
        create_and_add_last_patch_as_layer_patch(latest_layer, latest_layer_commit)
      else
        create_and_add_last_patch_as_latest_patch(latest_layer, latest_layer_commit)
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

    def archive_commitfile_path
      build_path filename '.commit'
    end

    def archive_commit
      File.read archive_commitfile_path
    end

    def create_arhive_with_owner_substitution!
      Dir.mktmpdir('change_archive_owner', build_path) do |tmpdir_path|
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

      File.write archive_commitfile_path, repo_latest_commit
    end

    def archive_exist?
      File.exist? archive_commitfile_path
    end

    def add_archive
      builder.docker.add_artifact archive_path, archive_filename, where_to_add, step: :prepare
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

    def add_patch(filename, step:)
      builder.docker.add_artifact(build_path(filename), filename, '/tmp', step: step)

      builder.docker.run(
        "zcat /tmp/#{filename} | #{sudo}git apply --whitespace=nowarn --directory=#{where_to_add}",
        "rm /tmp/#{filename}",
        step: step
      )
    end

    def create_patch!(from, filename, commitfile_path)
      atomizer << build_path(filename)
      atomizer << commitfile_path

      repo.git_bare "diff --binary #{from}..#{repo_latest_commit}#{" --relative=#{cwd}" if cwd} #{paths(true)} | gzip > #{build_path filename}"
      File.write commitfile_path, repo_latest_commit
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
      File.read layer_commitfile_path(layer) if File.exist? layer_commitfile_path(layer)
    end

    def layers
      Dir.glob(layer_commitfile_path('*')).map { |path| Integer(path.gsub(/.*_(\d+)\.commit$/, '\\1')) }.sort
    end

    def create_layer_patch!(from, layer)
      create_patch! from, layer_patch_filename(layer), layer_commitfile_path(layer)
    end

    def add_layer_patch(layer)
      add_patch layer_patch_filename(layer), step: :build
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
      File.read latest_commitfile_path if File.exist? latest_commitfile_path
    end

    def create_latest_patch!(from)
      create_patch! from, latest_patch_filename, latest_commitfile_path
    end

    def add_latest_patch
      add_patch latest_patch_filename, step: :setup
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
