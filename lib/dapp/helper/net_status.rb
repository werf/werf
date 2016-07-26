module Dapp
  module Helper
    # NetStatus
    module NetStatus
      def self.message(exception)
        net_status = exception.net_status.net_status_normalize(context: exception.net_status.delete(:context))
        net_status[:message] || [net_status[:error], net_status[:code]].compact.join(': ')
      end
    end
  end # Helper
end # Dapp
