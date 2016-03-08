module Dapper
  # Main class that does all stuff
  class Builder
    include Chefify
    include Centos7
    include CascadeTagging
    include Filelock

    class << self
      def default_opts
        @default_opts ||= {}
      end

      def process_directory(path, pattern = '*')
        dappfiles_paths = pattern.split('-').instance_eval { count.downto(1).map { |n| slice(0, n).join('-') } }.map { |p| Dir.glob(File.join(path, p, 'Dappfile')) }.find(&:any?) || []

        dappfiles_paths.map { |dappfile_path| process_file(dappfile_path, app_filter: pattern).builded_apps }.flatten
      end

      def process_file(dappfile_path, app_filter: '*')
        new(dappfile_path: dappfile_path, app_filter: app_filter) do |builder|
          builder.log "Processing application #{builder.name} (#{dappfile_path})"

          # indent all subsequent messages
          builder.indent_log

          # eval instructions from file
          builder.instance_eval File.read(dappfile_path), dappfile_path

          # commit atomizers
          builder.commit_atomizers!
        end
      end
    end

    def log(message)
      puts ' ' * opts[:log_indent] + ' * ' + message if opts[:log_verbose] || !opts[:log_quiet]
    end

    def shellout(*args, log_verbose: false, **kwargs)
      kwargs[:live_stream] = STDOUT if log_verbose && opts[:log_verbose]
      Mixlib::ShellOut.new(*args, :timeout => 3600, **kwargs).run_command.tap(&:error!)
    end

    def home_path(*path)
      path.compact.inject(Pathname.new(opts[:home_path]), &:+).expand_path.to_s
    end

    def build_path(*path)
      path.compact.inject(Pathname.new(opts[:build_path]), &:+).expand_path.tap do |p|
        FileUtils.mkdir_p p.parent
      end.to_s
    end

    def initialize(**options)
      opts.merge! self.class.default_opts
      opts.merge! options

      # default log indentation
      opts[:log_indent] = 0

      # basename
      if opts[:name]
        opts[:basename] = opts[:name]
        opts[:name] = nil
      elsif opts[:dappfile_path]
        opts[:basename] ||= Pathname.new(opts[:dappfile_path]).expand_path.parent.basename
      end

      # home path
      opts[:home_path] ||= Pathname.new(opts[:dappfile_path] || 'fakedir').parent.expand_path.to_s

      # build path
      opts[:build_path] = opts[:build_dir] ? opts[:build_dir] : home_path('build')
      opts[:build_path] = build_path opts[:basename] if opts[:shared_build_dir]

      # home branch
      @home_branch = shellout("git -C #{home_path} rev-parse --abbrev-ref HEAD").stdout.strip

      # atomizers
      @atomizers = []

      # account builded apps
      @builded_apps = []

      lock do
        yield self
      end
    end

    def indent_log
      opts[:log_indent] += 1
    end

    attr_reader :home_branch

    def builded_apps
      @builded_apps.dup
    end

    def docker
      docker_stack.last
    end

    def scope(&blk)
      stack_settings(&blk)
    end

    def app(name)
      log "Processing #{self.name}-#{name}"

      name = "#{opts[:name]}-#{name}" if opts[:name]

      stack_settings name: name, log_indent: opts[:log_indent] + 1 do
        docker.name name

        yield
      end
    end

    def name
      [opts[:basename], opts[:name]].compact.join '-'
    end

    def add_artifact_from_git(url, where_to_add, branch: 'master', ssh_key_path: nil, **kwargs)
      log "Adding artifact from git (#{url} to #{where_to_add}, branch: #{branch})"

      # extract git repo name from url
      repo_name = url.gsub(%r{.*?([^\/ ]+)\.git}, '\\1')

      # clone or fetch
      repo = GitRepo::Remote.new(self, repo_name, url: url, ssh_key_path: ssh_key_path)
      repo.fetch!(branch)

      # add artifact
      artifact = GitArtifact.new(self, repo, where_to_add, flush_cache: opts[:flush_cache], branch: branch, **kwargs)
      artifact.add_multilayer!
    end

    def build(**_kwargs)
      # check app name
      unless !opts[:app_filter] || File.fnmatch("#{opts[:app_filter]}*", name)
        log "Skipped (does not match filter: #{opts[:app_filter]})!"
        return false
      end

      # build image
      log 'Building'
      image_id = docker.build

      # apply cascade tagging
      tag_cascade image_id

      # push to registry
      if opts[:docker_registry]
        log 'Pushing to registry'
        docker.push name: name, registry: opts[:docker_registry]
      end

      # count it
      @builded_apps << name

      image_id
    end

    def tag(image_id, name: nil, tag: nil, registry: nil)
      return unless name && tag

      new = { name: name, tag: tag, registry: registry }
      docker.tag image_id, new
    end

    def register_atomizer(atomizer)
      atomizers << atomizer
    end

    def commit_atomizers!
      atomizers.each(&:commit!)
    end

    protected

    attr_reader :atomizers

    def opts
      opts_stack.last
    end

    def opts_stack
      @opts_stack ||= [{}]
    end

    def docker_stack
      @docker_stack ||= [Docker.new(self)]
    end

    def stack_settings(**options)
      opts_stack.push opts.merge(options).dup
      docker_stack.push docker.dup

      yield
    ensure
      docker_stack.pop
      opts_stack.pop
    end

    def lock(**kwargs, &block)
      filelock(build_path("#{home_branch}.lock"), error_message: "Application #{opts[:basename]} (#{home_branch}) in use! Try again later.", **kwargs, &block)
    end
  end
end
