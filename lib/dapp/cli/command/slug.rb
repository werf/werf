module Dapp
  class CLI
    module Command
      class Slug < Base
        banner <<BANNER.freeze
Usage:

  dapp slug STRING

Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          str = self.class.required_argument(self, 'string')
          run_dapp_command(nil, options: cli_options) do |dapp|
            dapp.slug([cli_arguments, str].flatten.join(' '))
          end
        end

        def log_running_time
          false
        end
      end
    end
  end
end
