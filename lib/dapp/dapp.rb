module Dapp
  class Dapp
    include Lock
    include GitArtifact
    include Dappfile
    include Chef

    include Logging::Base
    include Logging::Process
    include Logging::I18n
    include Logging::Paint

    include SshAgent
    include Helper::Sha256
    include Helper::Trivia
    include Helper::Tar

    include Deps::Gitartifact
    include Deps::Base

    include Shellout::Base

    attr_reader :options

    def initialize(options: {})
      @options = options
      Logging::Paint.initialize(options[:log_color])
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
          path.basename
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

    def path(*path)
      @path ||= expand_path(dappfile_path)
      make_path(@path, *path)
    end

    def tmp_base_dir
      File.expand_path(options[:tmp_dir_prefix] || '/tmp')
    end

    def build_path(*path)
      @build_path ||= begin
        if options[:build_dir]
          Pathname.new(options[:build_dir])
        else
          path('.dapp_build')
        end.expand_path.tap(&:mkpath)
      end
      make_path(@build_path, *path)
    end

    def local_git_artifact_exclude_paths(&blk)
      super do |exclude_paths|
        build_path_relpath = Pathname.new(build_path).subpath_of(File.dirname(git_path))
        exclude_paths << build_path_relpath.to_s if build_path_relpath

        yield exclude_paths if block_given?
      end
    end

    def stage_cache
      "dimgstage-#{name}"
    end

    def stage_dapp_label
      name
    end
  end # Dapp
end # Dapp
