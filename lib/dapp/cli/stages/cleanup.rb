require 'mixlib/cli'

module Dapp
  class CLI
    class Stages
      # stages cleanup subcommand
      class Cleanup < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages cleanup [options] [APPS PATTERN ...] [REPO]

    APPS PATTERN                Applications to process [default: *].

Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self)
          Project.new(cli_options: config, apps_patterns: cli_arguments).stages_cleanup(repo)
        end
      end
    end
  end
end
