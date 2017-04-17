module Dapp
  module Helper
    module NetStatus
      class << self
        def net_status(exception)
          exception.net_status.net_status_normalize(context: exception.net_status.delete(:context))
        end

        def message(exception)
          net_status = net_status(exception)
          net_status[:message] || begin
            data = net_status[:data].reject {|k, _| k == :backtrace}
            data = nil if data.empty?
            [net_status[:error], [net_status[:context], net_status[:code]].compact.join('.'), data].compact.join(': ')
          end
        end

        def before_error_message(exception)
          (net_status(exception)[:data][:before_error_messages] || []).join("\n")
        end
      end
    end
  end # Helper
end # Dapp
