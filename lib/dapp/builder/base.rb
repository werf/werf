module Dapp
  module Builder
    class Base
      include CommonHelper
      include Dapp::Builder::Centos7
      include Dapp::Builder::Ubuntu1404
      include Dapp::Builder::Ubuntu1604

      attr_reader :docker
      attr_reader :conf
      attr_reader :opts

      STAGES_DEPENDENCIES = {
          prepare: nil,
          infra_install: :prepare,
          sources_1: :infra_install,
          infra_setup: :sources_1,
          app_install: :infra_setup,
          sources_2: :app_install,
          app_setup: :sources_2,
          sources_3: :app_setup,
          sources_4: :sources_3
      }.freeze

      STAGES_DEPENDENCIES.each do |stage, dependence|
        define_method :"#{stage}_from" do
          send(:"#{dependence}_key") unless dependence.nil?
        end

        define_method :"#{stage}_image_name" do
          "dapp:#{send(:"#{stage}_key")}"
        end

        define_method :"#{stage}!" do
          docker.build_image!(image: send(stage), name: send(:"#{stage}_image_name"))
        end

        define_method :"#{stage}?" do
          docker.image_exist?(send("#{stage}_image_name"))
        end

        define_method stage do
          raise
        end
      end

      def initialize(docker:, conf:, opts:)
        @docker = docker
        @conf = conf
        @opts = opts
      end

      def run
        if prepare?
          prepare!
          infra_install!
          sources_1!
          infra_setup!
          app_install!
          app_setup!
        elsif infra_install?
          infra_install!
          sources_1!
          infra_setup!
          app_install!
          app_setup!
        elsif infra_setup?
          infra_setup!
          app_install!
          sources_2!
          app_setup!
        elsif app_install?
          app_install!
          sources_2!
          app_setup!
        elsif app_setup?
          app_setup!
          sources_3!
          sources_4!
        end
      end


      def prepare
        prepare_image
      end

      def prepare_key
        prepare_image.signature
      end

      def prepare_from
        conf[:from]
      end

      def prepare_image
        @prepare_image ||= begin
          image_method = :"from_#{conf[:from].to_s.split(/[:.]/).join}"
          raise "unsupported docker image '#{conf[:from]}'" unless respond_to?(image_method)
          send(image_method).tap do |image|
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
        if dependence_file?
          app_install_from
        else
          sha256([app_install_from, dependence_file, dependency_file_regex])
        end
      end

      def dependence_file
        @dependence_file ||= begin
          file_path = Dir[build_path('*')].detect {|x| x =~ dependency_file_regex }
          File.read(file_path) unless file_path.nil?
        end
      end

      def dependence_file?
        !dependence_file.nil?
      end

      def dependency_file_regex
        /.*\/(Gemfile|composer.json|requirement_file.txt)$/
      end


      def app_setup_key
        if app_setup_file?
          app_setup_from
        else
          sha256([app_setup_from, app_setup_file])
        end
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


      def home_branch #FIXME
        'master'
      end

      def make_local_git_artifact(cfg)
        repo = GitRepo::Own.new(self)
        GitArtifact.new(self, repo, cfg[:where_to_add],
                        flush_cache: opts[:flush_cache],
                        branch: home_branch)
        repo.fetch!(cfg[:branch])

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

      def sources_1_image
        @sources_1_image ||= Image.new(from: sources_1_from)
      end

      def sources_1
        git_artifact_list.each {|ga| ga.add_multilayer! sources_1_image}

        sources_1_image.build_opts!(volume: "#{build_path}:#{container_build_path}:ro")

        sources_1_image
      end

      def sources_1_key
        hashsum [sources_1_from, *git_artifact_list.map(&:signature)]
      end

      def sources_2
        sources_1
      end

      def sources_2_key
        sources_1_key
      end

      def sources_3
        sources_1
      end

      def sources_3_key
        sources_1_key
      end

      def sources_4
        sources_1
      end

      def sources_4_key
        sources_1_key
      end
    end # Base
  end # Builder
end # Dapp
