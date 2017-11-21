module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class Base < Base
          def run_dapp_command(run_command, options: {}, log_running_time: true)
            super(run_command, options: options.merge(verbose: true), log_running_time: log_running_time)
          end
        end
      end
    end
  end
end
