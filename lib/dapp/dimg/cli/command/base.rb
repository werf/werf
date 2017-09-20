module Dapp::Dimg::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments))
      end

      def run_dapp_command(run_method, *args)
        super(run_method, *args) do |dapp|
          begin
            dapp.host_docker_login
            yield dapp if run_method.nil? && block_given?
          ensure
            dapp.terminate
          end
        end
      end

      def run_method
        class_to_lowercase
      end
    end
  end
end
