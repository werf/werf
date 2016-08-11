module Dapp
  module Builder
    # Base
    class Base
      attr_reader :application

      def initialize(application)
        @application = application
      end

      def infra_install?
        false
      end

      def infra_install(_image)
        raise
      end

      def infra_install_checksum
        raise
      end

      def infra_setup?
        false
      end

      def infra_setup(_image)
        raise
      end

      def infra_setup_checksum
        raise
      end

      def install?
        false
      end

      def install(_image)
        raise
      end

      def install_checksum
        raise
      end

      def setup?
        false
      end

      def setup(_image)
        raise
      end

      def setup_checksum
        raise
      end

      def chef_cookbooks(_image)
      end

      def chef_cookbooks_checksum
        []
      end
    end # Base
  end # Builder
end # Dapp
