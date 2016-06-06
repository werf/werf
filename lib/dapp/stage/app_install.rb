module Dapp
  module Stage
    class AppInstall < Base
      def image
        super do |image|
          build.app_install_do(image)
        end
      end

      def signature
        hashsum build.stages[:source_1].signature
      end
    end # AppInstall
  end # Stage
end # Dapp
