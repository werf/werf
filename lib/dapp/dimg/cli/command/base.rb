module Dapp::Dimg::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        ::Dapp::Dapp.new(options: cli_options(dimgs_patterns: cli_arguments)).public_send(run_method)
      end

      def run_method
        class_to_lowercase
      end
    end
  end
end
