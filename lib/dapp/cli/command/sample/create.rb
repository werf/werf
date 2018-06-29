module Dapp
  class CLI
    module Command
      class Sample < ::Dapp::CLI
        class Create < Base
          banner <<BANNER.freeze
Usage:

  dapp sample create SAMPLE_NAME [options]

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            sample_name = self.class.required_argument(self, 'sample_name')
            run_dapp_command(nil, options: cli_options) do |dapp|
              dapp.sample_create(sample_name)
            end
          end
        end
      end
    end
  end
end
