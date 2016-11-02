module Dapp
  module Builder
    # Base
    class None < Base
      def before_dimg_should_be_built_check
      end

      def before_install(_image)
      end

      def before_install_checksum
      end

      def before_setup(_image)
      end

      def before_setup_checksum
      end

      def install(_image)
      end

      def install_checksum
      end

      def setup(_image)
      end

      def setup_checksum
      end

      def build_artifact(_image)
      end

      def build_artifact_checksum
      end

      def chef_cookbooks(_image)
      end

      def chef_cookbooks_checksum
      end
    end # Base
  end # Builder
end # Dapp
