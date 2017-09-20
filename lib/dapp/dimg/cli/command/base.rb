module Dapp::Dimg::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments))
      end

      def run_dapp_command(run_method, *args)
        super(nil, *args) do |dapp|
          begin
            dapp.host_docker_login
            if block_given?
              yield dapp
            elsif run_method.nil?
              dapp.public_send(run_method)
            end
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
