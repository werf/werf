module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class FlushRepo < Base
          banner <<BANNER.freeze
Usage:

  dapp dimg stages flush repo [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            repo = self.class.required_argument(self, 'repo')
            ::Dapp::Dapp.new(options: cli_options(dimgs_patterns: cli_arguments)).stages_flush_repo(repo)
          end
        end
      end
    end
  end
end
