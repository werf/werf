module Dapp::Dimg::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments))
      end

      def run_dapp_command(run_method, options: {}, log_running_time: true)
        super(nil, options: options, log_running_time: log_running_time) do |dapp|
          begin
            dapp.host_docker_login
            if block_given?
              yield dapp
            elsif !run_method.nil?
              dapp.public_send(run_method)
            end
          ensure
            dapp.terminate
          end
        end
      end
    end
  end
end
