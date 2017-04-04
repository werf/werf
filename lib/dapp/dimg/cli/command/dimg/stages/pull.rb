module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class Pull < Base
          banner <<BANNER.freeze
Usage:

  dapp dimg stages pull [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          option :pull_all_stages,
                 long: '--all',
                 boolean: true

          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            repo = self.class.required_argument(self)
            ::Dapp::Dapp.new(options: cli_options(dimgs_patterns: cli_arguments)).stages_pull(repo)
          end
        end
      end
    end
  end
end
