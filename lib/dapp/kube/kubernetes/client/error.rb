module Dapp
  module Kube
    module Kubernetes::Client::Error
      class Base < Kubernetes::Error::Base
      end

      class Default < Kubernetes::Error::Default
      end

      class Timeout < Default; end
      class ConnectionRefused < Default; end
      class BadConfig < Default; end

      class NotFound < Base
        def initialize(**net_status)
          super({code: :not_found}.merge(net_status))
        end
      end

      module Pod
        class NotFound < Kubernetes::Client::Error::NotFound ; end

        class ContainerCreating < Kubernetes::Client::Error::Base
          def initialize(**net_status)
            super({code: :container_creating}.merge(net_status))
          end
        end

        class PodInitializing < Kubernetes::Client::Error::Base
          def initialize(**net_status)
            super({code: :pod_initializing}.merge(net_status))
          end
        end
      end # Pod
    end # Kubernetes::Client::Error
  end # Kube
end # Dapp
