module Dapp
  module Stage
    class AppInstall < Base
      def name
        :app_install
      end

      def image
        super do |image|
          builder.app_install_do(image)
        end
      end

      def signature
        hashsum builder.stages[:source_1].signature
      end
    end # AppInstall
  end # Stage
end # Dapp
