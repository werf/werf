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
          option :registry_username,
                 long: '--registry-username USERNAME'

          option :registry_password,
                 long: '--registry-password PASSWORD'

          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            repo = self.class.required_argument(self, 'repo')
            run_dapp_command(:stages_flush_repo, options: cli_options(dimgs_patterns: cli_arguments, repo: repo))
          end
        end
      end
    end
  end
end
