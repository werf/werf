module Dapp
  module Stage
    class InfraInstall < Base
      def image
        super do |image|
          builder.infra_install_do(image)
        end
      end

      def signature
        builder.infra_install_signature_do
      end
    end # InfraInstall
  end # Stage
end # Dapp
