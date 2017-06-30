module Dapp
  module Error
    # Shellout
    class Shellout < Base
      def initialize(net_status={})
        super( {code: :shell_command_failed}.merge(net_status) )
      end
    end
  end
end
