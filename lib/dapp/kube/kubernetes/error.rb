module Dapp
  module Kube
    module Kubernetes::Error
      class Base < ::Dapp::Error::Base
        def initialize(**net_status)
          super(**net_status, context: :kubernetes)
        end
      end

      class Default < Base
        include ::Dapp::Error::Mod::User
      end
    end # Kubernetes::Error
  end # Kube
end # Dapp
