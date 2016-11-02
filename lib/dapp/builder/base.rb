module Dapp
  module Builder
    # Base
    class Base
      attr_reader :dimg

      def initialize(dimg)
        @dimg = dimg
      end

      def before_dimg_should_be_built_check
      end

      def before_install?
        false
      end

      def before_install(_image)
        raise
      end

      def before_install_checksum
        raise
      end

      def before_setup?
        false
      end

      def before_setup(_image)
        raise
      end

      def before_setup_checksum
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

      def build_artifact?
        false
      end

      def build_artifact(_image)
        raise
      end

      def build_artifact_checksum
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
