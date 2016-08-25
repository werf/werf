module Dapp
  # Project
  class Project
    include Lock
    include Dappfile
    include Paint
    include Command::Common
    include Command::Run
    include Command::Build
    include Command::Bp
    include Command::Push
    include Command::Spush
    include Command::List
    include Command::Stages::CleanupLocal
    include Command::Stages::CleanupRepo
    include Command::Stages::FlushLocal
    include Command::Stages::FlushRepo
    include Command::Stages::Common
    include Command::Cleanup
    include Logging::Base
    include Logging::Process

    include SshAgent
    include Helper::I18n
    include Helper::Shellout
    include Helper::Paint
    include Helper::Sha256

    attr_reader :cli_options
    attr_reader :apps_patterns

    def initialize(cli_options: {}, apps_patterns: nil)
      @cli_options = cli_options
      @apps_patterns = apps_patterns || []
      @apps_patterns << '*' unless @apps_patterns.any?

      paint_initialize
      Helper::I18n.initialize
    end

    def name
      @name ||= begin
        shellout!("git -C #{path} config --get remote.origin.url").stdout.strip.split('/').last[/.*(?=.git)/]
      rescue Error::Shellout
        File.basename(path)
      end
    end

    def path
      @path ||= begin
        dappfile_path = dappfiles.first
        if File.basename(expand_path(dappfile_path, 2)) == '.dapps'
          expand_path(dappfile_path, 3)
        else
          expand_path(dappfile_path)
        end
      end
    end

    def build_path
      @build_path ||= begin
        if cli_options[:build_dir]
          Pathname.new(cli_options[:build_dir])
        else
          Pathname.new(path).join('.dapps_build')
        end.expand_path.tap(&:mkpath)
      end
    end

    def cache_format
      "dappstage-#{name}-%{application_name}"
    end

    def stage_dapp_label_format
      '%{application_name}'
    end
  end # Project
end # Dapp
