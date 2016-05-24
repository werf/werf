module Dapp
  module Builder
    class Base
      attr_reader :conf

      def initialize(conf)
        @conf = conf
      end

      def prepare
        raise
      end

      def prepare_key
        raise
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
