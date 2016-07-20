module Dapp
  module Builder
    # Base
    class Base
      attr_reader :application

      def initialize(application)
        @application = application
      end

      def infra_install(_image)
        fail
      end

      def infra_install_checksum
        fail
      end

      def infra_setup(_image)
        fail
      end

      def infra_setup_checksum
        fail
      end

      def app_install(_image)
        fail
      end

      def app_install_checksum
        fail
      end

      def app_setup(_image)
        fail
      end

      def app_setup_checksum
        fail
      end
    end # Base
  end # Builder
end # Dapp
