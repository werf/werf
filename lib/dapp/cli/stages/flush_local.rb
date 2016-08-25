module Dapp
  class CLI
    class Stages
      # stages flush local subcommand
      class FlushLocal < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages flush local [options] [APPS PATTERN ...]

    APPS PATTERN                Applications to process [default: *].

Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Project.new(cli_options: config, apps_patterns: cli_arguments).stages_flush_local
        end
      end
    end
  end
end
