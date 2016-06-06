module Dapp
  module Stage
    class InfraSetup < Base
      def name
        :infra_setup
      end

      def image
        super do |image|
          builder.infra_setup_do(image)
        end
      end

      def signature
        hashsum builder.stages[:source_2].signature
      end
    end # InfraSetup
  end # Stage
end # Dapp
