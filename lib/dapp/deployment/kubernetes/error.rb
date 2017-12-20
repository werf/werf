module Dapp
  module Deployment
    module Kubernetes::Error
      class Default < ::Dapp::Deployment::Error::Kubernetes
        def initialize(**net_status)
          super(**net_status, context: :kubernetes)
        end
      end

      class NotFound < Default
        def initialize(**net_status)
          super({code: :not_found}.merge(net_status))
        end
      end

      class Timeout < Default; end
      class ConnectionRefused < Default; end
    end
  end
end
