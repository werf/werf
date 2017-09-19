module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretRegenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret regenerate [SECRET_VALUES_FILE_PATH ...] [options]

Options:
BANNER

      option :old_secret_key,
             long: '--old-secret-key KEY',
             description: 'Old secret key',
             required: true

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(nil, options: cli_options) do |dapp|
          dapp.public_send(run_method, *cli_arguments)
        end
      end
    end
  end
end
