module Dapp
  module Builder
    # Base
    class Base
      attr_reader :application

      def initialize(application)
        @application = application
      end

      def infra_install(_image)
        raise
      end

      def infra_install_checksum
        raise
      end

      def infra_setup(_image)
        raise
      end

      def infra_setup_checksum
        raise
      end

      def install(_image)
        raise
      end

      def install_checksum
        raise
      end

      def setup(_image)
        raise
      end

      def setup_checksum
        raise
      end
    end # Base
  end # Builder
end # Dapp
