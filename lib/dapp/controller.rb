module Dapp
  class Controller
    include CommonHelper

    attr_reader :cli_options, :patterns

    def initialize(cli_options:, patterns: nil)
      @cli_options = cli_options
      @cli_options[:log_indent] = 0

      @patterns = patterns || []
      @patterns << '*' unless @patterns.any?

      build_confs
    end

    def build
      @build_confs.each { |build_conf|
        log build_conf._name
        log_build_time(!cli_options[:show_only]) do
          with_log_indent { Application.new(config: build_conf, cli_options: cli_options).build! }
        end
      }
    end

    def list
      @build_confs.each do |build_conf|
        log build_conf._name
      end
    end

    def show
      @build_confs.each do |build_conf|
        log build_conf._name
        with_log_indent { log JSON.pretty_generate(build_conf.to_h) }
      end
    end

    def push(repo)
      raise "Several applications isn't available for push command!" unless @build_confs.one?
      log @build_confs.first._name
      with_log_indent { Application.new(config: @build_confs.first, cli_options: cli_options, ignore_git_fetch: true).export!(repo) }
    end

    def smartpush(repo_prefix)
      @build_confs.each do |build_conf|
        log build_conf._name
        repo = File.join(repo_prefix, build_conf._name)
        with_log_indent { Application.new(config: build_conf, cli_options: cli_options, ignore_git_fetch: true).export!(repo) }
      end
    end

    def flush_build_cache
      @build_confs.each do |build_conf|
        log build_conf._name
        app = Application.new(config: build_conf, cli_options: cli_options, ignore_git_fetch: true)
        FileUtils.rm_rf app.build_cache_path
      end
    end

    def self.flush_stage_cache
      shellout('docker rmi $(docker images --format="{{.Repository}}:{{.Tag}}" dapp)')
      shellout('docker rmi $(docker images -f "dangling=true" -q)')
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
        log "Processing dappfile '#{dappfile_path}'"
        conf.instance_eval File.read(dappfile_path), dappfile_path
      end
      config._apps.select { |app| app_filters.any? { |pattern| File.fnmatch(pattern, app._name) } }
    end

    def dappfile_path
      @dappfile_path ||= File.join [cli_options[:dir], 'Dappfile'].compact
    end

    def dapps_path
      @dapps_path ||= File.join [cli_options[:dir], '.dapps'].compact
    end

    def log_build_time(log, &blk)
      time = run_time(&blk)
      log("build time: #{time.round(2)}") if log
    end

    def run_time
      start = Time.now
      yield
      Time.now - start
    end
  end # Controller
end # Dapp
