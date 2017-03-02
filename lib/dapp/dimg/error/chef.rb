module Dapp
  module Dimg
    module Error
      class Chef < Base
        def initialize(**net_status)
          super(context: 'chef', **net_status)
        end
      end
    end
  end
end
