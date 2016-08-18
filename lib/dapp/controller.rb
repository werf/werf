module Dapp
  # Controller
  class Controller
    include Helper::Log
    include Helper::I18n
    include Helper::Shellout

    attr_reader :cli_options
    attr_reader :patterns

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
      build_configs.each { |config| puts config._name }
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
        containers_flush(basename)
        with_subquery(%(docker images --format="{{.Repository}}:{{.Tag}}" #{basename}-dappstage)) { |ids| shellout!(%(docker rmi #{ids.join(' ')})) }
      end
    end

    def stages_cleanup(repo)
      repo_apps = repo_apps(repo)
      build_configs.map(&:_basename).uniq.each do |basename|
        log(basename)
        containers_flush(basename)
        apps, stages = project_images(basename).partition { |_, image_id| repo_apps.values.include?(image_id) }
        apps = apps.to_h
        stages = stages.to_h
        apps.each do |_, aiid|
          iid = aiid
          until (iid = image_parent(iid)).empty?
            stages.delete_if { |_, siid| siid == iid }
          end
        end
        shellout!(%(docker rmi #{stages.keys.join(' ')})) unless stages.keys.empty?
      end
    end

    def cleanup
      build_configs.map(&:_basename).uniq.each do |basename|
        log(basename)
        containers_flush(basename)
        with_subquery(%(docker images -f "dangling=true" -f "label=dapp=#{basename}" -q)) { |ids| shellout!(%(docker rmi #{ids.join(' ')})) }
        with_subquery(%(docker images --format '{{if ne "#{basename}-dappstage" .Repository }}{{.ID}}{{ end }}' -f "label=dapp=#{basename}")) do |ids|
          shellout!(%(docker rmi #{ids.join(' ')}))
        end # FIXME: negative filter is not currently supported by the Docker CLI
      end
    end

    private

    def repo_apps(repo)
      registry = DockerRegistry.new(repo)
      raise Error::Registry, :no_such_app unless registry.repo_exist?
      registry.repo_apps
    end

    def containers_flush(basename)
      with_subquery(%(docker ps -a -f "label=dapp" -f "name=#{basename}" -q)) do |ids|
        shellout!(%(docker rm -f #{ids.join(' ')}))
      end
    end

    def project_images(basename)
      shellout!(%(docker images --format "{{.Repository}}:{{.Tag}};{{.ID}}" --no-trunc #{basename}-dappstage)).stdout.lines.map do |line|
        line.strip.split(';')
      end.to_h
    end

    def image_parent(image_id)
      shellout!(%(docker inspect -f {{.Parent}} #{image_id})).stdout.strip
    end

    def with_subquery(query)
      return if (res = shellout!(query).stdout.strip.lines.map(&:strip)).empty?
      yield(res)
    end

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
      config = Config::Main.new(dappfile_path: dappfile_path, controller: self) do |conf|
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
