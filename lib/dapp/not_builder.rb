module Dapp
  # FIXME rename Controller
  class NotBuilder
    include CommonHelper

    attr_reader :opts, :patterns

    class << self
      def flush_stage_cache
        # FIXME
        # shellout('docker rmi $(docker images --format="{{.Repository}}:{{.Tag}}" registry)')
        # shellout('docker rmi $(docker images -f "dangling=true" -q)')
        # 
        3.times do
          image_names = %w(none dapp)
          image_names.each { |image_name| shellout("docker rmi -f $(docker images | grep \"#{image_name}\" | awk \"{print \$3}\")") }
        end
      end
    end

    def initialize(cli_options:, patterns: nil)
      @opts = cli_options
      @opts[:log_indent] = 0

      @patterns = patterns || []
      @patterns << '*' unless @patterns.any?

      build_confs
    end

    def build
      @build_confs.each { |build_conf|
        log build_conf.name
        with_log_indent { Application.new(conf: build_conf, opts: opts).build_and_fixate! }
      }
    end

    def list
      @build_confs.each do |build_conf|
        log build_conf.name
      end
    end

    def show
      @build_confs.each do |build_conf|
        log build_conf.name
        with_log_indent { log JSON.pretty_generate(build_conf.to_h) }
      end
    end

    def push(repo)
      raise "Several applications isn't available for push command!" unless @build_confs.one?
      log @build_confs.first.name
      with_log_indent { Application.new(conf: @build_confs.first, opts: opts, ignore_git_fetch: true).push!(repo) }
    end

    def smartpush(repo_prefix)
      @build_confs.each do |build_conf|
        log build_conf.name
        tag_name = File.join(repo_prefix, build_conf.name)
        with_log_indent { Application.new(conf: build_conf, opts: opts, ignore_git_fetch: true).push!(tag_name) }
      end
    end

    def flush_build_cache
      @build_confs.each do |build_conf|
        log build_conf.name
        app = Application.new(conf: build_conf, opts: opts, ignore_git_fetch: true)
        FileUtils.rm_rf app.build_cache_path
      end
    end

    private

    def build_confs
      @build_confs ||= begin
        dappfiles = []
        if File.exist? dappfile_path
          dappfiles << dappfile_path
        elsif File.exist? dapps_path
          dappfiles += Dir.glob(File.join([dapps_path, '*', 'Dappfile'].compact))
        end
        dappfiles.flatten.uniq!
        apps = dappfiles.map { |dappfile| apps(dappfile, app_filters: patterns) }.flatten

        if apps.empty?
          STDERR.puts "Error: No such app: '#{patterns.join(', ')}' in #{dappfile_path}"
          exit 1
        else
          apps
        end
      end
    end

    def apps(dappfile_path, app_filters:)
      config = Config::Main.new(dappfile_path: dappfile_path) do |conf|
        conf.log "Processing dappfile '#{dappfile_path}'"
        conf.instance_eval File.read(dappfile_path), dappfile_path
      end
      config.apps.select { |app| app_filters.any? { |pattern| File.fnmatch(pattern, app.name) } }
    end

    def dappfile_path
      @dappfile_path ||= File.join [opts[:dir], 'Dappfile'].compact
    end

    def dapps_path
      @dapps_path ||= File.join [opts[:dir], '.dapps'].compact
    end
  end # NotBuilder
end # Dapp
