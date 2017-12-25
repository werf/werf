module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class Base < Base
          def run_dapp_command(run_command, options: {}, log_running_time: true, **extra_options)
            super(run_command, options: options.merge(verbose: true), log_running_time: log_running_time, **extra_options)
          end
        end
      end
    end
  end
end
