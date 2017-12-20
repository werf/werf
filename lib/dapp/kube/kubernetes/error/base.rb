module Dapp
  module Kube
    module Kubernetes::Error
      class Base < ::Dapp::Error::Base
        def initialize(**net_status)
          super(**net_status, context: :kubernetes)
        end
      end

      class Default ::Dapp::Error::Default
        def initialize(**net_status)
          super(**net_status, context: :kubernetes)
        end
      end
    end # Kubernetes::Error
  end # Kube
end # Dapp
