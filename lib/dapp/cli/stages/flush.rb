require 'mixlib/cli'

module Dapp
  class CLI
    class Stages
      # Flush stages subcommand
      class Flush < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages flush [options] [APPS PATTERN ...]

    APPS PATTERN                Applications to process [default: *].

Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Project.new(cli_options: config, apps_patterns: cli_arguments).stages_flush
        end
      end
    end
  end
end
