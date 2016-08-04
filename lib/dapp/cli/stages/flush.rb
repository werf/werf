require 'mixlib/cli'

module Dapp
  class CLI
    class Stages
      # Flush stages subcommand
      class Flush < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages flush
Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.new(cli_options: config, patterns: cli_arguments).stages_flush
        end
      end
    end
  end
end
