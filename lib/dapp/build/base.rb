module Dapp
  module Build
    class Base
      include CommonHelper
      include Dapp::Filelock

      attr_accessor :docker
      attr_reader :conf
      attr_reader :opts
      attr_reader :home_branch
      attr_reader :starter_stage
      attr_reader :builder
      attr_reader :docker_atomizer

      def initialize(conf:, opts:, builder:)
        @conf = conf
        @opts = opts
        @builder = builder

        # default log indentation
        opts[:log_indent] = 0

        opts[:build_path] = opts[:build_dir] ? opts[:build_dir] : home_path('build')
        opts[:build_path] = build_path opts[:basename] if opts[:shared_build_dir]

	# FIXME ambigous variable name
        @starter_stage = Stage::Source5.new(self)
        @docker = Dapp::Docker.new(socket: opts[:docker_socket], build: self)
      end

      def run
        starter_stage.do_build
        builder.commit_atomizers!
      end

      def signature
        starter_stage.signature
      end

      def git_artifact_list
        [local_git_artifact, *remote_git_artifact_list].compact
      end

      def local_git_artifact
        @local_git_artifact ||= begin
          cfg = (conf[:git_artifact] || {})[:local]
          make_local_git_artifact(cfg) if cfg
        end
      end

      def remote_git_artifact_list
        @remote_git_artifact_list ||= Array((conf[:git_artifact] || {})[:remote])
            .map(&method(:make_local_git_artifact)) # FIXME make_remote_git_artifact
      end

      def home_path(*path)
        path.compact.inject(Pathname.new(conf[:home_path]), &:+).expand_path
      end

      def build_path(*path)
        path.compact.inject(Pathname.new(opts[:build_path]), &:+).expand_path.tap do |p|
          FileUtils.mkdir_p p.parent
        end
      end

      def container_build_path(*path)
        path.compact.inject(Pathname.new('/.build'), &:+).expand_path
      end

      # FIXME wtf in base class????
      def chef_path(*path)
        path.compact.inject(build_path('chef'), &:+).expand_path.tap do |p|
          FileUtils.mkdir_p p.parent
        end
      end

      # FIXME wtf in base class????
      def container_chef_path(*path)
        path.compact.inject(container_build_path('chef'), &:+).expand_path
      end

      protected

      def infra_install_do(_image)
        raise
      end

      # FIXME commands -> check_sum
      def infra_install_commands
        raise
      end


      def infra_setup_do(_image)
        raise
      end

      # FIXME commands -> check_sum
      def infra_setup_commands
        raise
      end


      def app_install_do(_image)
        raise
      end

      # FIXME commands -> check_sum
      def app_install_commands
        raise
      end


      def app_setup_do(_image)
        raise
      end

      # FIXME commands -> signature
      def app_setup_commands
        raise
      end

      def make_local_git_artifact(cfg)
        repo = GitRepo::Own.new(self)
        GitArtifact.new(self, repo, cfg[:where_to_add],
                        flush_cache: opts[:flush_cache],
                        branch: cfg[:branch])
      end

      def make_remote_git_artifact(cfg)
        repo_name = cfg[:url].gsub(%r{.*?([^\/ ]+)\.git}, '\\1')
        repo = GitRepo::Remote.new(self, repo_name,
                                   url: cfg[:url],
                                   ssh_key_path: ssh_key_path)
        repo.fetch!(cfg[:branch])
        GitArtifact.new(self, repo, cfg[:where_to_add],
                        flush_cache: opts[:flush_cache],
                        branch: cfg[:branch])
      end
    end # Base
  end # Builder
end # Dapp
