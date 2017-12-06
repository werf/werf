module Dapp
  module Kube
    module Error
      class Base < ::Dapp::Error::Base
        def initialize(**net_status)
          super(**net_status, context: :kube)
        end
      end
    end
  end
end
