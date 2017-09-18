module Dapp::Kube::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(run_method, options: cli_options)
      end

      def run_method
        :"kube_#{class_to_lowercase}"
      end
    end
  end
end
