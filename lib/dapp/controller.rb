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

    def spush(repo)
      raise Error::Controller, code: :spush_command_unexpected_apps_number unless build_configs.one?
      Application.new(config: build_configs.first, cli_options: cli_options, ignore_git_fetch: true).tap do |app|
        app.export!(repo, format: '%{repo}:%{tag}')
      end
    end

    def push(repo)
      build_configs.each do |config|
        log_step(config._name)
        with_log_indent do
          Application.new(config: config, cli_options: cli_options, ignore_git_fetch: true).tap do |app|
            app.export!(repo, format: '%{repo}:%{app_name}-%{tag}')
          end
        end
      end
    end

    def stages_flush
      build_configs.map(&:_basename).uniq.each do |basename|
        log(basename)
        shellout(%{docker rmi $(docker images --format="{{.Repository}}:{{.Tag}}" #{basename}-dappstage)})
      end
    end

    def cleanup
      build_configs.map(&:_basename).uniq.each do |basename|
        log(basename)
        shellout('docker rm -f $(docker ps -a -f "label=dapp-meta" -q)')
        shellout(%{docker rmi $(docker images -f "dangling=true" -f "label=dapp=#{basename}" -q)})
        shellout(%{docker rmi $(docker images --format '{{if ne "#{basename}-dappstage" .Repository }}{{.ID}} {{ end }}' -f "label=dapp=#{basename}" | sed '/^$/d')}) # FIXME: negative filter is not currently supported by the Docker CLI
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
          raise Error::Controller, code: :no_such_app, data: { patterns: patterns.join(', ') } if apps.empty?
        end
      end
    end

    def apps(dappfile_path, app_filters:)
      config = Config::Main.new(dappfile_path: dappfile_path) do |conf|
        begin
          conf.instance_eval File.read(dappfile_path), dappfile_path
        rescue SyntaxError, StandardError => e
          message = case e
                    when ArgumentError then "#{e.backtrace.first[/`.*'$/]}: #{e.message}"
                    else e.message
                    end
          raise Error::Dappfile, code: :incorrect, data: { error: e.class.name, path: dappfile_path, message: message }
        end
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
