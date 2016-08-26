module Dapp
  class CLI
    class Stages
      # stages push subcommand
      class Push < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages push [options] [APP PATTERN] REPO

    APP PATTERN                 Applications to process [default: *].

Options:
BANNER
        option :lock_timeout,
               long: '--lock-timeout TIMEOUT',
               description: 'Redefine resource locking timeout (in seconds)',
               proc: ->(v) { v.to_i }

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self)
          Project.new(cli_options: config, apps_patterns: cli_arguments).stages_push(repo)
        end
      end
    end
  end
end
