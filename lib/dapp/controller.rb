module Dapp
  # Controller
  class Controller
    include Helper::Log
    include Helper::Shellout

    attr_reader :cli_options, :patterns

    def initialize(cli_options: {}, patterns: nil)
      @cli_options = cli_options
      @cli_options[:log_indent] = 0

      @patterns = patterns || []
      @patterns << '*' unless @patterns.any?

      paint_initialize
      Helper::I18n.initialize
      build_confs
    end

    def run(docker_options, command)
      raise Error::Controller, code: :run_command_unexpected_apps_number unless build_confs.one?
      Application.new(config: build_confs.first, cli_options: cli_options, ignore_git_fetch: true).run(docker_options, command)
    end

    def build
      build_confs.each do |build_conf|
        log_step(build_conf._name)
        with_log_indent { Application.new(config: build_conf, cli_options: cli_options).build! }
      end
    end

    def list
      build_confs.each do |build_conf|
        log(build_conf._name)
      end
    end

    def push(repo)
      raise Error::Controller, code: :push_command_unexpected_apps_number unless build_confs.one?
      Application.new(config: build_confs.first, cli_options: cli_options, ignore_git_fetch: true).export!(repo)
    end

    def smartpush(repo_prefix)
      build_confs.each do |build_conf|
        log_step(build_conf._name)
        repo = File.join(repo_prefix, build_conf._name)
        with_log_indent { Application.new(config: build_conf, cli_options: cli_options, ignore_git_fetch: true).export!(repo) }
      end
    end

    def flush_metadata
      build_confs.each do |build_conf|
        log(build_conf._name)
        app = Application.new(config: build_conf, cli_options: cli_options, ignore_git_fetch: true)
        FileUtils.rm_rf app.metadata_path
      end
    end

    def self.flush_stages
      shellout('docker rmi $(docker images --format="{{.Repository}}:{{.Tag}}" dapp)')
      shellout('docker rmi $(docker images -f "dangling=true" -q)')
    end

    private

    def build_confs
      @build_confs ||= begin
        if File.exist? dappfile_path
          dappfiles = dappfile_path
        elsif (dappfiles = dapps_dappfiles_pathes).empty? && (dappfiles = search_dappfile_up).nil?
          raise Error::Controller, code: :dappfile_not_found
        end
        Array(dappfiles).map { |dappfile| apps(dappfile, app_filters: patterns) }.flatten.tap do |apps|
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
      File.join [cli_options[:dir], 'Dappfile'].compact
    end

    def dapps_dappfiles_pathes
      Dir.glob(File.join([cli_options[:dir], '.dapps', '*', 'Dappfile'].compact))
    end

    def search_dappfile_up
      cdir = Pathname(File.expand_path(cli_options[:dir] || Dir.pwd))
      until (cdir = cdir.parent).root?
        next unless (path = cdir.join('Dappfile')).exist?
        return path.to_s
      end
    end

    def paint_initialize
      Paint.mode = case cli_options[:log_color]
                   when 'auto' then STDOUT.tty? ? 8 : 0
                   when 'on'   then 8
                   when 'off'  then 0
                   else raise
                   end
    end
  end # Controller
end # Dapp
