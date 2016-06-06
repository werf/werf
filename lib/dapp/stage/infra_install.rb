module Dapp
  module Stage
    class InfraInstall < Base
      def image
        super do |image|
          build.infra_install_do(image)
        end
      end

      def signature
        image.signature
      end
    end # InfraInstall
  end # Stage
end # Dapp
