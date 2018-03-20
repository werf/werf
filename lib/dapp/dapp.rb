module Dapp
  class Dapp
    include Lock
    include GitArtifact
    include Dappfile
    include Chef
    include DappConfig
    include OptionTags
    include Slug

    include Command::Common
    include Command::Slug

    include Logging::Base
    include Logging::Process
    include Logging::I18n
    include Logging::Paint

    include SshAgent
    include Sentry
    include Ruby2Go

    include Helper::Sha256
    extend  Helper::Trivia
    include Helper::Trivia
    include Helper::Tar
    include Helper::Url

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

      @_call_before_terminate = []

      ruby2go_init
    end

    def options
      self.class.options
    end

    def settings
      @settings ||= begin
        settings_path = File.join(self.class.home_dir, "settings.toml")

        if File.exists? settings_path
          TomlRB.load_file(settings_path)
        else
          {}
        end
      end
    end

    def name
      @name ||= begin
        n = begin
          if (name = options[:name])
            name
          elsif git_own_repo_exist?
            if git_url
              repo_name = git_url.split('/').last
              repo_name = repo_name[/.*(?=\.git)/] if repo_name.end_with? '.git'
              repo_name
            else
              File.basename(File.dirname(git_own_repo.path)).to_s
            end
          else
            path.basename.to_s
          end
        end
        consistent_uniq_slugify(n)
      end
    end

    def git_url
      return unless git_own_repo_exist?
      git_own_repo.remote_origin_url
    rescue Dimg::Error::Rugged => e
      return if e.net_status[:code] == :git_repository_without_remote_url
      raise
    end

    def git_own_repo_exist?
      !git_own_repo.nil?
    end

    def git_own_repo
      @git_own_repo ||= Dimg::GitRepo::Own.new(self)
    rescue Dimg::Error::Rugged => e
      raise unless e.net_status[:code] == :local_git_repository_does_not_exist
      nil
    end

    def work_dir
      File.expand_path(options[:dir] || Dir.pwd)
    end

    def path(*path)
      @path ||= make_path(work_dir)
      make_path(@path, *path)
    end

    def tmp_base_dir
      self.class.tmp_base_dir
    end

    def build_dir
      @build_dir ||= begin
        if option_build_dir
          Pathname.new(option_build_dir)
        else
          dir = File.join(self.class.home_dir, "builds", self.name)
          Pathname.new(dir)
        end.expand_path.tap(&:mkpath)
      end
    end

    def build_path(*path)
      make_path(build_dir, *path)
    end

    def local_git_artifact_exclude_paths(&blk)
      super do |exclude_paths|
        build_path_relpath = Pathname.new(build_path).subpath_of(File.dirname(git_own_repo.path))
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
      @_call_before_terminate.each {|on_terminate| on_terminate.call(self)}
      FileUtils.rmtree(host_docker_tmp_config_dir)
    end

    def host_docker
      self.class.host_docker
    end

    def host_docker_tmp_config_dir
      self.class.host_docker_tmp_config_dir
    end

    def try_host_docker_login
      return unless option_repo
      validate_repo_name!(option_repo)
      host_docker_login(option_repo)
    end

    def host_docker_login(repo)
      return unless self.class.options_with_docker_credentials?
      username, password = self.class.docker_credentials
      shellout!("#{host_docker} login -u '#{username}' -p '#{password}' '#{repo}'")
    end

    class << self
      def home_dir
        File.join(Dir.home, ".dapp")
      end

      def options
        @options ||= {}
      end

      def host_docker
        @host_docker ||= begin
          min_docker_minor_version = Gem::Version.new('1.10')
          unless host_docker_minor_version > min_docker_minor_version
            raise Error::Dapp, code: :docker_version, data: { min_version: min_docker_minor_version.to_s,
                                                              version:     host_docker_minor_version.to_s }
          end

          [].tap do |cmd|
            cmd << host_docker_bin
            cmd << "--config #{host_docker_config_dir}"
          end.join(' ')
        end
      end

      def host_docker_bin
        raise Error::Dapp, code: :docker_not_found if (res = shellout('which docker')).exitstatus.nonzero?
        res.stdout.strip
      end

      def host_docker_minor_version
        Gem::Version.new(shellout!("#{host_docker_bin} --version").stdout.strip[/\d+\.\d+/])
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
        !docker_credentials.nil?
      end

      def docker_credentials
        if options.key?(:registry_username) && options.key?(:registry_password)
          [options[:registry_username], options[:registry_password]]
        elsif ENV.key?('CI_JOB_TOKEN')
          ['gitlab-ci-token', ENV['CI_JOB_TOKEN']]
        end
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
