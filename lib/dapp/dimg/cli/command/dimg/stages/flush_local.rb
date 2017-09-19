module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class FlushLocal < Base
          banner <<BANNER.freeze
Usage:

  dapp dimg stages flush local [options] [DIMG ...]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            run_dapp_command(:stages_flush_local, options: cli_options(dimgs_patterns: cli_arguments))
          end
        end
      end
    end
  end
end
