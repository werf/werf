module Dapp
  module Stage
    class InfraSetup < Base
      def image
        super do |image|
          builder.infra_setup_do(image)
        end
      end

      def signature
        hashsum builder.stages[:sources_2].signature
      end
    end # InfraSetup
  end # Stage
end # Dapp
