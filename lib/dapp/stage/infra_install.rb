module Dapp
  module Stage
    class InfraInstall < Base
      def name
        :infra_install
      end

      def image
        super do |image|
          builder.infra_install_do(image)
        end
      end

      def signature
        image.signature
      end
    end # InfraInstall
  end # Stage
end # Dapp
