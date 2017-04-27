module Dapp
  module Deployment
    module Kubernetes::Error
      class Base < ::Dapp::Deployment::Error::Kubernetes
        def initialize(**net_status)
          super(**net_status, context: :kubernetes)
        end
      end

      class NotFound < Base
        def initialize(**net_status)
          super({code: :not_found}.merge(net_status))
        end
      end

      class Timeout < Base; end
    end
  end
end
