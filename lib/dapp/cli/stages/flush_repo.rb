module Dapp
  class CLI
    class Stages
      # stages flush repo subcommand
      class FlushRepo < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages flush repo [options] [APPS PATTERN ...] REPO

    APPS PATTERN                Applications to process [default: *].

Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self)
          Project.new(cli_options: config, apps_patterns: cli_arguments).stages_flush_repo(repo)
        end
      end
    end
  end
end
