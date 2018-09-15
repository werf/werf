module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretEdit < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret edit FILE_PATH [options]

Options:
BANNER

      option :values,
             long: '--values',
             description: 'Edit secret values file',
             default: false

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        file_path = self.class.required_argument(self, 'FILE_PATH')
        run_dapp_command(nil, options: cli_options) do |dapp|
          dapp.public_send(run_method, file_path)
        end
      end

      def log_running_time
        false
      end
    end
  end
end
