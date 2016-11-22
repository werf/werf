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
    include Command::Tag
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
    include SystemShellout

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
        if git_path
          begin
            system_shellout("#{git_bin} -C #{path} config --get remote.origin.url").stdout.strip.split('/').last[/.*(?=.git)/]
          rescue
            File.basename(path)
          end
        else
          File.basename(path)
        end
      end
    end

    def git_path
      defined?(@git_path) ? @git_path : begin
        dot_git_path = search_file_upward('.git')
        @git_path = dot_git_path ? Pathname.new(File.dirname(dot_git_path)) : nil
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
          Pathname.new(path).join('.dapp_build')
        end.expand_path.tap(&:mkpath)
      end
    end

    def stage_cache
      "dimgstage-#{name}"
    end

    def stage_dapp_label
      name
    end

    def dev_mode?
      !!cli_options[:dev]
    end
  end # Project
end # Dapp
