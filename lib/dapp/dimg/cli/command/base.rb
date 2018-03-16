module Dapp::Dimg::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      option :build_dir,
             long: "--build-dir PATH",
             description: "Directory where build cache stored ($HOME/.dapp/builds/<dapp name> by default)."

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments))
      end
    end
  end
end
