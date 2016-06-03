module Dapp
  module Builder
    class Base
      include CommonHelper
      include Dapp::Builder::Stages
      include Dapp::Filelock

      attr_reader :docker
      attr_reader :conf
      attr_reader :opts
      attr_reader :home_branch

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
        sources_3_build! if sources_3_build?
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


      def sources_key
        hashsum git_artifact_list.map(&:signature)
      end

      def sources_1_key
        sources_key
      end

      def sources_2_key
        if sources_1_image_exist?
          sources_2_dependence_key
        else
          sources_key
        end
      end

      def sources_2_image_exist?
        sources_1_image_exist? or docker.image_exist?(sources_2_image_name)
      end

      def sources_3_key
        if sources_2_image_exist?
          sources_3_dependence_key
        else
          sources_key
        end
      end

      def sources_3_image_exist?
        sources_2_image_exist? or docker.image_exist?(sources_3_image_name)
      end

      def sources_4_key
        #TODO: split sources_3 into period-layer + latest patch scheme
      end


      def sources_1_build_image?
        (not sources_1_image_exist?) and begin
          infra_setup_build_image? or
            app_install_build_image? or
              app_setup_build_image?
        end
      end

      def sources_2_build_image?
        (not sources_2_image_exist?) and app_setup_build_image?
      end


      def sources(image)
        image.tap do
          git_artifact_list.each {|ga| ga.add_multilayer! image}
          image.build_opts!(volume: "#{build_path}:#{container_build_path}:ro")
        end
      end

      def sources_1
        super do
          sources sources_1_image
        end
      end

      def sources_2
        super do
          sources sources_2_image
        end
      end

      def sources_3
        super do
          sources sources_3_image
        end
      end

      def sources_4
        #TODO: split sources_3 into period-layer + latest patch scheme
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
