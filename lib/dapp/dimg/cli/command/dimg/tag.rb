module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Tag < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg tag [options] [DIMG] TAG
  
    DIMG                        Dapp image to process [default: *].

Options:
BANNER

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          tag = self.class.required_argument(self, 'tag')
          run_dapp_command(nil, options: cli_options(dimgs_patterns: cli_arguments)) do |dapp|
            dapp.public_send(run_method, tag)
          end
        end
      end
    end
  end
end
