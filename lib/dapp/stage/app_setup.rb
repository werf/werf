module Dapp
  module Stage
    class AppSetup < Base
      def name
        :app_setup
      end

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
