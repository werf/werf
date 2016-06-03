module Dapp
  module Builder
    class Base
      include CommonHelper
      include Dapp::Filelock

      attr_reader :docker
      attr_reader :conf
      attr_reader :opts
      attr_reader :home_branch
      attr_reader :stages

      def initialize(docker:, conf:, opts:)
        @docker = docker
        @conf = conf
        @opts = opts

        opts[:home_path] ||= Pathname.new(opts[:dappfile_path] || 'fakedir').parent.expand_path.to_s
        opts[:build_path] = opts[:build_dir] ? opts[:build_dir] : home_path('build')
        opts[:build_path] = build_path opts[:basename] if opts[:shared_build_dir]

        @home_branch = shellout("git -C #{home_path} rev-parse --abbrev-ref HEAD").stdout.strip
        @atomizers = []
        @builded_apps = []

        @stages = {
          prepare: Dapp::Stage::Prepare.new(self),
          infra_install: Dapp::Stage::InfraInstall.new(self),
          source_1_archive: Dapp::Stage::Source1Archive.new(self),
          source_1: Dapp::Stage::Source1.new(self),
          app_install: Dapp::Stage::AppInstall.new(self),
          source_2:  Dapp::Stage::Source2.new(self),
          infra_setup: Dapp::Stage::InfraSetup.new(self),
          source_3: Dapp::Stage::Source3.new(self),
          app_setup: Dapp::Stage::AppSetup.new(self),
          source_4: Dapp::Stage::Source4.new(self),
          source_5: Dapp::Stage::Source5.new(self),
        }.tap {|stages|
          stages.values.reduce {|prev, stage|
            prev.next = stage
            stage.prev = prev
            stage
          }
        }

        lock do
          yield self
        end if block_given?
      end

      def lock(**kwargs, &blk)
        filelock(build_path("#{home_branch}.lock"),
                 error_message: "Application #{opts[:basename]} " +
                                "(#{home_branch}) in use! Try again later.",
                 **kwargs, &blk)
      end

      def run
        stages.values.last.build
        commit_atomizers!
      end

      def infra_install_do(_image)
        raise
      end

      def infra_install_signature_do
        raise
      end


      def infra_setup_do(_image)
        raise
      end

      def infra_setup_signature_do
        raise
      end


      def app_install_do(_image)
        raise
      end

      def app_install_signature_do
        raise
      end


      def app_setup_do(_image)
        raise
      end

      def app_setup_signature_do
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

      def local_git_artifact
        @local_git_artifact ||= begin
          cfg = (conf[:git_artifact] || {})[:local]
          make_local_git_artifact(cfg) if cfg
        end
      end

      def remote_git_artifact_list
        @remote_git_artifact_list ||= Array((conf[:git_artifact] || {})[:remote])
                                      .map(&method(:make_local_git_artifact))
      end

      def git_artifact_list
        [local_git_artifact, *remote_git_artifact_list].compact
      end


      def home_path(*path)
        path.compact.inject(Pathname.new(opts[:home_path]), &:+).expand_path
      end

      def build_path(*path)
        path.compact.inject(Pathname.new(opts[:build_path]), &:+).expand_path.tap do |p|
          FileUtils.mkdir_p p.parent
        end
      end

      def container_build_path(*path)
        path.compact.inject(Pathname.new('/.build'), &:+).expand_path
      end

      def chef_path(*path)
        path.compact.inject(build_path('chef'), &:+).expand_path.tap do |p|
          FileUtils.mkdir_p p.parent
        end
      end

      def container_chef_path(*path)
        path.compact.inject(container_build_path('chef'), &:+).expand_path
      end

      def register_atomizer(atomizer)
        atomizers << atomizer
      end

      def commit_atomizers!
        atomizers.each(&:commit!)
      end

      protected

      attr_reader :atomizers
    end # Base
  end # Builder
end # Dapp
