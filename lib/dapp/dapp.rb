module Dapp
  class Dapp
    include Lock
    include GitArtifact
    include Dappfile
    include Chef
    include DappConfig

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
      Logging::I18n.initialize
      validate_config_options!
      Logging::Paint.initialize(option_color)
    end

    def name
      @name ||= begin
        if git_url
          repo_name = git_url.split('/').last
          repo_name = repo_name[/.*(?=\.git)/] if repo_name.end_with? '.git'
          repo_name
        elsif git_path
          File.basename(File.dirname(git_path)).to_s
        else
          path.basename.to_s
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
        if option_build_dir
          Pathname.new(option_build_dir)
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

    def host_docker
      self.class.host_docker
    end

    def self.host_docker
      @host_docker ||= begin
        raise Error::Dapp, code: :docker_not_found if (res = shellout('which docker')).exitstatus.nonzero?
        docker_bin = res.stdout.strip

        current_docker_version = shellout!("#{docker_bin} --version").stdout.strip
        required_docker_version = '1.10.0'

        if Gem::Version.new(required_docker_version) >= Gem::Version.new(current_docker_version[/(\d+\.)+\d+/])
          raise Error::Dapp, code: :docker_version, data: { version: required_docker_version }
        end

        [].tap do |cmd|
          cmd << docker_bin
          cmd << "--config #{ENV['DAPP_DOCKER_CONFIG']}" if ENV.key?('DAPP_DOCKER_CONFIG')
        end.join(' ')
      end
    end
  end # Dapp
end # Dapp
