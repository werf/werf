module Dapp
  module Error
    class Base < NetStatus::Exception
      def initialize(net_status = {})
        super(net_status.merge(context: context))
      end

      private

      def context
        ['error', class_name].join('.')
      end

      def class_name
        self.class.to_s.split('::').last.downcase
      end
    end
  end
end
