module Dapp
  # Application
  class Application
    include Helper::Log
    include Helper::Shellout
    include Helper::I18n
    include Helper::Sha256
    include Logging
    include GitArtifact
    include Path
    include Tags
    include Dapp::Filelock

    attr_reader :config
    attr_reader :cli_options
    attr_reader :ignore_git_fetch

    def initialize(config:, cli_options:, ignore_git_fetch: false)
      @config = config
      @cli_options = cli_options

      @build_path = cli_options[:build_dir] || home_path('build')
      @build_cache_path = cli_options[:build_cache_dir] || home_path('build_cache')

      @last_stage = Build::Stage::Source5.new(self)
      @ignore_git_fetch = ignore_git_fetch
    end

    def build!
      last_stage.build!
      last_stage.save_in_cache!
    end

    def export!(repo)
      raise Error::Application, code: :application_not_built unless last_stage.image.tagged? || dry_run?

      tags.each do |tag|
        image_name = [repo, tag].join(':')
        if dry_run?
          log_state(image_name, state: t(code: 'state.push'), styles: { status: :success })
        else
          log_process(image_name, process: t(code: 'status.process.pushing')) do
            last_stage.image.export!(image_name, log_verbose: log_verbose?, log_time: log_time?, force: cli_options[:force])
          end
        end
      end
    end

    def builder
      @builder ||= Builder.const_get(config._builder.capitalize).new(self)
    end

    protected

    attr_reader :last_stage
  end # Application
end # Dapp
