module Dapp
  class CLI
    class Stages
      # stages cleanup local subcommand
      class CleanupLocal < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages cleanup local [options] [APPS PATTERN ...] REPO

    APPS PATTERN                Applications to process [default: *].

Options:
BANNER
        option :proper_cache_version,
               long: '--improper-cache-version',
               boolean: true

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self)
          Project.new(cli_options: config, apps_patterns: cli_arguments).stages_cleanup_local(repo)
        end
      end
    end
  end
end
