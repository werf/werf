module Dapp
  module Stage
    class AppSetup < Base
      def image
        super do |image|
          build.app_setup_do(image)
        end
      end

      def signature
        hashsum build.stages[:source_3].signature
      end
    end # AppSetup
  end # Stage
end # Dapp
