module Dapp
  module Builder
    class Chef < Base
      # Error
      class Error < ::Dapp::Error::Base
        def initialize(**net_status)
          super(context: :chef, **net_status)
        end
      end
    end # Chef
  end # Builder
end # Dapp
