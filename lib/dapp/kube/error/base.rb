module Dapp
  module Kube
    module Error
      class Base < ::Dapp::Error::Base
        def initialize(net_status = {})
          super({context: 'kube'}.merge(net_status))
        end
      end
    end
  end
end
