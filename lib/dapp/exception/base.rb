module Dapp
  module Exception
    # Base
    class Base < NetStatus::Exception
      def initialize(net_status = {})
        super(net_status.merge(context: self.class.to_s.split('::').last.downcase))
      end
    end
  end
end
