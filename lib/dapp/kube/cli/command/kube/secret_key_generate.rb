module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretKeyGenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret key generate

Options:
BANNER
      def run_dapp_command(run_method, options: {}, log_running_time: false)
        super(run_method, options: options, log_running_time: log_running_time)
      end
    end
  end
end
