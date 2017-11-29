module Dapp
  class Dapp
    include Lock
    include GitArtifact
    include Dappfile
    include Chef
    include DappConfig
    include OptionTags

    include Command::Common

    include Logging::Base
    include Logging::Process
    include Logging::I18n
    include Logging::Paint

    include SshAgent
    include Helper::Sha256
    extend Helper::Trivia
    include Helper::Trivia
    include Helper::Tar

    include Deps::Toolchain
    include Deps::Gitartifact
    include Deps::Base

    include Shellout::Base

    attr_reader :options

    def initialize(options: {})
      self.class.options.merge!(options)
      Logging::I18n.initialize
      validate_config_options!
      Logging::Paint.initialize(option_color)
    end

    def options
      self.class.options
    end

    def name
      @name ||= begin
        if name = options[:name]
          name
        elsif git_url
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
      self.class.tmp_base_dir
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

    def terminate
      FileUtils.rmtree(host_docker_tmp_config_dir)
    end

    def host_docker
      self.class.host_docker
    end

    def host_docker_tmp_config_dir
      self.class.host_docker_tmp_config_dir
    end

    def host_docker_login
      return unless option_repo

      validate_repo_name!(option_repo)

      login = proc {|u, p| shellout!("#{host_docker} login -u '#{u}' -p '#{p}' '#{option_repo}'")}
      if options.key?(:registry_username) && options.key?(:registry_password)
        login.call(options[:registry_username], options[:registry_password])
      elsif ENV.key?('CI_JOB_TOKEN')
        login.call('gitlab-ci-token', ENV['CI_JOB_TOKEN'])
      end
    end

    class << self
      def options
        @options ||= {}
      end

      def host_docker
        @host_docker ||= begin
          raise Error::Dapp, code: :docker_not_found if (res = shellout('which docker')).exitstatus.nonzero?
          docker_bin = res.stdout.strip

          current_docker_version = shellout!("#{docker_bin} --version").stdout.strip
          required_min_docker_version = '1.10.0'
          required_max_docker_version = '17.09.0'

          if Gem::Version.new(required_min_docker_version) > Gem::Version.new(current_docker_version[/(\d+\.)+\d+/]) ||
              Gem::Version.new(required_max_docker_version) < Gem::Version.new(current_docker_version[/(\d+\.)+\d+/])
            raise Error::Dapp, code: :docker_version, data: { min_version: required_min_docker_version,
                                                              max_version: required_max_docker_version,
                                                              version: current_docker_version[/(\d+\.)+\d+/] }
          end

          [].tap do |cmd|
            cmd << docker_bin
            cmd << "--config #{host_docker_config_dir}"
          end.join(' ')
        end
      end

      def host_docker_config_dir
        if options_with_docker_credentials? && !options[:repo].nil?
          host_docker_tmp_config_dir
        elsif ENV.key?('DAPP_DOCKER_CONFIG')
          ENV['DAPP_DOCKER_CONFIG']
        else
          File.join(Dir.home, '.docker')
        end
      end

      def options_with_docker_credentials?
        (options.key?(:registry_username) && options.key?(:registry_password)) || ENV.key?('CI_JOB_TOKEN')
      end

      def host_docker_tmp_config_dir
        @host_docker_tmp_config_dir ||= Dir.mktmpdir('dapp-', tmp_base_dir)
      end

      def tmp_base_dir
        File.expand_path(options[:tmp_dir_prefix] || '/tmp')
      end
    end
  end # Dapp
end # Dapp
