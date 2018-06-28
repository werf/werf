module Dapp
  class CLI
    module Command
      class Example < ::Dapp::CLI
        class Create < Base
          banner <<BANNER.freeze
Usage:

  dapp example create EXAMPLE_NAME [options]

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            example_name = self.class.required_argument(self, 'example_name')
            run_dapp_command(nil, options: cli_options) do |dapp|
              dapp.example_create(example_name)
            end
          end
        end
      end
    end
  end
end
