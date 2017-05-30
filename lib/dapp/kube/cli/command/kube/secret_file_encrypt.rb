module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretFileEncrypt < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret file encrypt FILE [options]

Options:
BANNER

      option :output_file_path,
             short: '-o OUTPUt_FILE_PATH',
             required: false

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        file_path = self.class.required_argument(self, 'file')
        ::Dapp::Dapp.new(options: cli_options).public_send(run_method, file_path)
      end
    end
  end
end
