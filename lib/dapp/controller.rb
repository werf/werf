module Dapp
  # Controller
  class Controller
    include Helper::Log
    include Helper::Shellout
    include Helper::I18n

    attr_reader :cli_options, :patterns

    def initialize(cli_options:, patterns: nil)
      @cli_options = cli_options
      @cli_options[:log_indent] = 0

      @patterns = patterns || []
      @patterns << '*' unless @patterns.any?

      i18n_initialize
      build_confs
    end

    def build
      @build_confs.each do |build_conf|
        log build_conf._name.base
        with_log_indent { Application.new(config: build_conf, cli_options: cli_options).build! }
      end
    end

    def list
      @build_confs.each do |build_conf|
        log build_conf._name.base
      end
    end

    def push(repo)
      raise Error::Controller, code: :one_application_expected unless @build_confs.one?
      Application.new(config: @build_confs.first, cli_options: cli_options, ignore_git_fetch: true).export!(repo)
    end

    def smartpush(repo_prefix)
      @build_confs.each do |build_conf|
        log build_conf._name.base
        repo = File.join(repo_prefix, build_conf._name)
        with_log_indent { Application.new(config: build_conf, cli_options: cli_options, ignore_git_fetch: true).export!(repo) }
      end
    end

    def flush_build_cache
      @build_confs.each do |build_conf|
        log build_conf._name.base
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
        else
          raise Error::Controller, code: :dappfile_not_found
        end
        dappfiles.flatten.uniq!
        dappfiles.map { |dappfile| apps(dappfile, app_filters: patterns) }.flatten.tap do |apps|
          raise Error::Controller, code: :no_such_app, data: { path: dappfile_path, patterns: patterns.join(', ') } if apps.empty?
        end
      end
    end

    def apps(dappfile_path, app_filters:)
      config = Config::Main.new(dappfile_path: dappfile_path) do |conf|
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
  end # Controller
end # Dapp
