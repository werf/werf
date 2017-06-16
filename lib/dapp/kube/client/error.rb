module Dapp
  module Kube
    module Client::Error
      class Base < ::Dapp::Kube::Error::Kubernetes
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
      class ConnectionRefused < Base; end
      class BadConfig < Base; end
    end
  end
end
