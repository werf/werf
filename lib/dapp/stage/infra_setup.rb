module Dapp
  module Stage
    class InfraSetup < Base
      def image
        super do |image|
          builder.infra_setup_do(image)
        end
      end

      def signature
        builder.infra_setup_signature_do
      end
    end # InfraSetup
  end # Stage
end # Dapp
