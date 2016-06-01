module Dapp
  module Builder
    class Base
      include CommonHelper
      include Dapp::Builder::Centos7
      include Dapp::Builder::Ubuntu1404
      include Dapp::Builder::Ubuntu1604
      include Dapp::Filelock

      attr_reader :docker
      attr_reader :conf
      attr_reader :opts
      attr_reader :home_branch

      STAGES_DEPENDENCIES = {
          prepare: nil,
          infra_install: :prepare,
          sources_1: :infra_install,
          infra_setup: :sources_1,
          app_install: :infra_setup,
          sources_2: :app_install,
          app_setup: :sources_2,
          sources_3: :app_setup,
          #sources_4: :sources_3
      }.freeze

      STAGES_DEPENDENCIES.each do |stage, dependence|
        define_method :"#{stage}_from" do
          send(:"#{dependence}_image_name") unless dependence.nil?
        end

        define_method :"#{stage}_image_name" do
          "dapp:#{send(:"#{stage}_key")}"
        end

        define_method(:"#{stage}_image") do
          instance_variable_get(:"@#{stage}_image") ||
            instance_variable_set(:"@#{stage}_image", Image.new(from: send(:"#{stage}_from")))
        end

        define_method :"#{stage}_image_exist?" do
          docker.image_exist?(send("#{stage}_image_name"))
        end

        define_method :"#{stage}_build?" do
          (send(:"#{dependence}_build?") if dependence) or send(:"#{stage}_build_image?")
        end

        define_method :"#{stage}_build_image?" do
          not send(:"#{stage}_image_exist?")
        end

        define_method :"#{stage}_build_image!" do
          image = send(stage)
          docker.build_image!(image: image, name: send(:"#{stage}_image_name")) if image
        end

        define_method :"#{stage}_build!" do
          send(:"#{dependence}_build!") if dependence and send(:"#{dependence}_build?")
          send(:"#{stage}_build_image!")
        end

        define_method stage do
          raise
        end
      end

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

      def prepare
        prepare_image
      end

      def prepare_key
        prepare_image.signature
      end

      def _image_method(from)
        :"from_#{from.to_s.split(/[:.]/).join}"
      end

      def prepare_from
        conf[:from].tap do |from|
          raise "unsupported docker image '#{from}'" unless respond_to?(_image_method(from))
        end
      end

      def prepare_image
        @prepare_image ||= begin
          send(_image_method(prepare_from)).tap do |image|
            image.build_options[:expose] = conf[:exposes] unless conf[:exposes].nil?
          end
        end
      end


      def infra_install_key
        infra_install_from
      end


      def infra_setup_key
        infra_setup_from
      end


      def app_install_key
        hashsum [app_install_from, dependency_file, dependency_file_regex]
      end

      def dependency_file
        @dependency_file ||= begin
          file_path = Dir[build_path('*')].detect {|x| x =~ dependency_file_regex }
          File.read(file_path) unless file_path.nil?
        end
      end

      def dependency_file?
        !dependency_file.nil?
      end

      def dependency_file_regex
        /.*\/(Gemfile|composer.json|requirement_file.txt)$/
      end


      def app_setup_key
        hashsum [app_setup_from, app_setup_file]
      end

      def app_setup_file
        @app_setuo_file ||= begin
          File.read(app_setup_file_path) if app_setup_file?
        end
      end

      def app_setup_file?
        File.exist?(app_setup_file_path)
      end

      def app_setup_file_path
        build_path('.app_setup')
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


      def sources_key(from)
        hashsum [from, *git_artifact_list.map(&:signature)]
      end

      def sources_1_key
        sources_key sources_1_from
      end

      def sources_2_key
        sources_key sources_2_from
      end

      def sources_3_key
        sources_key sources_3_from
      end

      def sources_4_key
        #TODO: split sources_3 into period-layer + latest patch scheme
      end


      def sources_1_build_image?
        not sources_1_exist? and begin
          infra_setup_build? or
            app_install_build? or
              app_setup_build?
        end
      end

      def sources_2_build_image?
        not sources_2_exist? and app_setup_build?
      end


      def sources(image)
        image.tap do
          git_artifact_list.each {|ga| ga.add_multilayer! image}
          image.build_opts!(volume: "#{build_path}:#{container_build_path}:ro")
        end
      end

      def sources_1
        sources sources_1_image
      end

      def sources_2
        sources sources_2_image
      end

      def sources_3
        sources sources_3_image
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
