module Dapp
  module Stage
    class AppSetup < Base
      def image
        super do |image|
          builder.app_setup_do(image)
        end
      end

      def signature
        hashsum builder.stages[:source_3].signature
      end
    end # AppSetup
  end # Stage
end # Dapp
