module Dapp
  module Kube
    module Kubernetes::Client::Error
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

      module Pod
        class NotFound < Kubernetes::Client::Error::NotFound ; end
      end # Pod
    end # Kubernetes::Client::Error
  end # Kube
end # Dapp
