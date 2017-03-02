module Dapp
  module Dimg
    module CLI
      class Stages
        # stages push subcommand
        class Push < Base
          banner <<BANNER.freeze
Version: #{::Dapp::VERSION}

Usage:
  dapp dimg stages push [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            repo = self.class.required_argument(self)
            Dapp.new(cli_options: config, dimgs_patterns: cli_arguments).stages_push(repo)
          end
        end
      end
    end
  end
end
