module Dapp
  module Error
    class Default < Base
      include Mod::User

      def initialize(net_status = {})
        super({ context: self.class.to_s.split('::').last.downcase }.merge(net_status))
      end
    end
  end
end
