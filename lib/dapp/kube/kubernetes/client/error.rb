module Dapp
  module Kube
    module Kubernetes::Client::Error
      class Default < ::Dapp::Kube::Error::Kubernetes
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
      class BadConfig < Default; end

      module Pod
        class NotFound < Kubernetes::Client::Error::NotFound ; end

        class ContainerCreating < Kubernetes::Client::Error::Default
          def initialize(**net_status)
            super({code: :container_creating}.merge(net_status))
          end
        end

        class PodInitializing < Kubernetes::Client::Error::Default
          def initialize(**net_status)
            super({code: :pod_initializing}.merge(net_status))
          end
        end
      end # Pod
    end # Kubernetes::Client::Error
  end # Kube
end # Dapp
