module Dapp
  module Builder
    class Base
      include CommonHelper

      attr_reader :docker
      attr_reader :conf

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

        define_method :"#{stage}_image" do
          "dapp:#{send(:"#{stage}_key")}"
        end

        define_method :"#{stage}!" do
          build_stage!(from: send(:"#{stage}_from"), stage: stage)
        end

        define_method :"#{stage}?" do
          @docker.image_exist?(send("#{stage}_image"))
        end

        define_method stage do
          raise
        end
      end

      # TODO Describe stages sequence with
      # TODO   ordering data
      # TODO Generate stages related methods
      # TODO   from that data

      def initialize(docker, conf)
        @docker = docker
        @conf = conf
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

      def build_stage!(from:, stage:)
        raise
      end


      def prepare
        # запуск shell-команд из conf
      end

      def prepare_from
        conf[:from]
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
          sha256([app_install_key, dependence_file, dependency_file_regex])
        end
      end

      def dependence_file
        file_path = Dir[File.join(config[:home_dir], '*')].detect {|x| x =~ dependency_file_regex }
        File.read(file_path) unless file_path.nil?
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
        File.read(File.join(config[:home_dir], '.app_setup')) if app_setup_file?
      end

      def app_setup_file?
        File.exist?(File.join(config[:home_dir], '.app_setup'))
      end


      def sources_1_key
        raise
      end


      def sources_2_key
        raise
      end


      def sources_3_key
        raise
      end


      def sources_4_key
        raise
      end
    end
  end
end
