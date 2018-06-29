module Dapp
  class CLI
    module Command
      class Sample < ::Dapp::CLI
        class List < Base
          banner <<BANNER.freeze
Usage:

  dapp sample list [options]

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            run_dapp_command(run_method, options: cli_options)
          end

          def log_running_time
            false
          end
        end
      end
    end
  end
end
