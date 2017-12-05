module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class CleanupRepo < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg cleanup repo [options] [DIMG ...] REPO

Options:
BANNER

        option :lock_timeout,
               long: '--lock-timeout TIMEOUT',
               description: 'Redefine resource locking timeout (in seconds)',
               proc: ->(v) { v.to_i }

        option :with_stages,
               long: '--with-stages',
               boolean: true

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self, 'repo')
          run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments, repo: repo))
        end
      end
    end
  end
end
