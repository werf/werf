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
        if git_url
          repo_name = git_url.split('/').last
          repo_name = repo_name[/.*(?=\.git)/] if repo_name.end_with? '.git'
          repo_name
        elsif git_path
          File.basename(File.dirname(git_path))
        else
          File.basename(path)
        end
      end
    end

    def git_url
      return unless git_config
      (git_config['remote "origin"'] || {})['url']
    end

    def git_config
      @git_config ||= begin
        IniFile.load(File.join(git_path, 'config')) if git_path
      end
    end

    def git_path
      defined?(@git_path) ? @git_path : begin
        @git_path = search_file_upward('.git')
      end
    end

    def path
      @path ||= expand_path(dappfile_path)
    end

    def cookbook_path
      File.join(path, '.dapp_chef')
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

    def system_files
      [dappfile_path, cookbook_path, build_path].map { |p| File.basename(p) }
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
