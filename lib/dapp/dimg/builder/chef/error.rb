module Dapp::Dimg::Builder
  class Chef < Base
    class Error < ::Dapp::Error::Base
      def initialize(**net_status)
        super(context: 'chef', **net_status)
      end
    end
  end # Chef
end # Dapp::Dimg::Builder
