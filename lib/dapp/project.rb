module Dapp
  # Project
  class Project
    include Lock
    include Dappfile

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
    include Command::Stages::Push
    include Command::Stages::Pull
    include Command::Stages::Common
    include Command::Cleanup
    include Command::Mrproper
    include Command::StageImage
    include Logging::Base
    include Logging::Process
    include Logging::I18n
    include Logging::Paint

    include SshAgent
    include Helper::Sha256
    include Helper::Trivia

    include Deps::Gitartifact
    include Deps::Base

    include Shellout::Base
    include Shellout::System

    attr_reader :cli_options
    attr_reader :dimgs_patterns

    def initialize(cli_options: {}, dimgs_patterns: nil)
      @cli_options = cli_options
      @dimgs_patterns = dimgs_patterns || []
      @dimgs_patterns << '*' unless @dimgs_patterns.any?

      Logging::Paint.initialize(cli_options[:log_color])
      Logging::I18n.initialize
    end

    def name
      @name ||= begin
        if File.exist? File.join(path, '.git')
          system_shellout("#{git_path} -C #{path} config --get remote.origin.url").stdout.strip.split('/').last[/.*(?=.git)/] rescue File.basename(path)
        else
          File.basename(path)
        end
      end
    end

    def path
      @path ||= expand_path(dappfile_path)
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

    def stage_cache
      "dimgstage-#{name}"
    end

    def stage_dapp_label
      name
    end
  end # Project
end # Dapp
