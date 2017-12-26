module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class FlushLocal < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg flush local [options] [DIMG ...]

Options:
BANNER

        option :with_stages,
               long: '--with-stages',
               boolean: true

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments, verbose: true))
        end
      end
    end
  end
end
