module Dapp
  class CLI
    module Command
      class Example < ::Dapp::CLI
        class List < Base
          banner <<BANNER.freeze
Usage:

  dapp example list [options]

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            run_dapp_command(run_method, options: cli_options)
          end
        end
      end
    end
  end
end
