module Dapp::Kube::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        ::Dapp::Dapp.new(options: cli_options).public_send(run_method)
      end

      def run_method
        :"kube_#{class_to_lowercase}"
      end
    end
  end
end
