module Dapp
  # Lock
  module Lock
    # Error
    module Error
      # Timeout
      class Timeout < ::Dapp::Error::Base
        def initialize(**net_status)
          super(context: :lock, **net_status)
        end
      end # Timeout
    end # Error
  end # Lock
end # Dapp
