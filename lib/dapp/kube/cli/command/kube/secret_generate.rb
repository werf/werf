module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretGenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret generate [FILE_PATH] [options]

Options:
BANNER

      option :output_file_path,
             short: '-o OUTPUT_FILE_PATH',
             required: false

      option :values,
             long: '--values',
             description: 'Decode secret values file',
             default: false

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        file_path = cli_arguments.empty? ? nil : cli_arguments.first
        run_dapp_command(nil, options: cli_options) do |dapp|
          dapp.public_send(run_method, file_path)
        end
      end
    end
  end
end
