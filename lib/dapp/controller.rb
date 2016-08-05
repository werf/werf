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
    end

    def run(docker_options, command)
      raise Error::Controller, code: :run_command_unexpected_apps_number unless build_configs.one?
      Application.new(config: build_configs.first, cli_options: cli_options, ignore_git_fetch: true).run(docker_options, command)
    end

    def build
      build_configs.each do |config|
        log_step(config._name)
        with_log_indent do
          Application.new(config: config, cli_options: cli_options).build!
        end
      end
    end

    def list
      build_configs.each { |config| log(config._name) }
    end

    def push(repo)
      raise Error::Controller, code: :push_command_unexpected_apps_number unless build_configs.one?
      Application.new(config: build_configs.first, cli_options: cli_options, ignore_git_fetch: true).export!(repo)
    end

    def smartpush(repo_prefix)
      build_configs.each do |config|
        log_step(config._name)
        repo = File.join(repo_prefix, config._name)
        with_log_indent { Application.new(config: config, cli_options: cli_options, ignore_git_fetch: true).export!(repo) }
      end
    end

    def stages_flush
      shellout(%{docker rmi $(docker images --format="{{.Repository}}:{{.Tag}}" #{build_configs.first._basename}-dappstage)})
    end

    def stages_cleanup
      shellout(%{docker rmi $(docker images -f "dangling=true" -f "label=dapp=#{build_configs.first._basename}" -q)})
    end

    def metadata_flush
      build_configs.each do |config|
        log(config._name)
        app = Application.new(config: config, cli_options: cli_options, ignore_git_fetch: true)
        FileUtils.rm_rf app.metadata_path
      end
    end

    private

    def build_configs
      @configs ||= begin
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
