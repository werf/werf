module Dapp
  module Error
    class Base < NetStatus::Exception
      def initialize(net_status = {})
        super({ context: self.class.to_s.split('::').last.downcase }.merge(net_status))
      end
    end
  end
end
