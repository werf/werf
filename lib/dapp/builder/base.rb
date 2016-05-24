module Dapp
  module Builder
    class Base
      attr_reader :conf

      def initialize(conf)
        @conf = conf
      end

      def run
=begin
Варианты запуска стадий

  prepare
  infra_install
  sources_1
  infra_setup
  app_install
  app_setup

  infra_install
  sources_1
  infra_setup
  app_install
  app_setup

  infra_setup
  app_install
  sources_2
  app_setup

  infra_setup
  app_install
  sources_2
  app_setup

  app_install
  sources_2
  app_setup

  app_setup
  sources_3
  sources_4
=end
      end

      def build_docker_image(from:, cmd: [], tag:)
        # запустить команды в новом контейнере через docker run
        # сделать docker commit
        # удалить контейнер
      end

      def prepare
        # запуск shell-команд из conf
      end

      def prepare_key
        # hash от shell-команд
      end

      def infra_install
        raise
      end

      def infra_install_key
        raise
      end

      def sources_1
        raise
      end

      def sources_1_key
        raise
      end

      def infra_setup
        raise
      end

      def infra_setup_key
        raise
      end

      def app_install
        raise
      end

      def app_install_key
        raise
      end

      def sources_2
        raise
      end

      def sources_2_key
        raise
      end

      def app_setup
        raise
      end

      def app_setup_key
        raise
      end

      def sources_3
        raise
      end

      def sources_3_key
        raise
      end

      def sources_4
        raise
      end

      def sources_4_key
        raise
      end
    end
  end
end
