module Dapp
  module Dimg
    module CLI
      class Stages
        # stages flush repo subcommand
        class FlushRepo < Base
          banner <<BANNER.freeze
Version: #{::Dapp::VERSION}

Usage:
  dapp dimg stages flush repo [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            repo = self.class.required_argument(self)
            Dapp.new(cli_options: config, dimgs_patterns: cli_arguments).stages_flush_repo(repo)
          end
        end
      end
    end
  end
end
