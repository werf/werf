require 'mixlib/cli'

module Dapp
  class CLI
    class Stages
      # Flush stages subcommand
      class Cleanup < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages cleanup
Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.new(cli_options: config, patterns: cli_arguments).stages_cleanup
        end
      end
    end
  end
end
