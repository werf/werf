require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      # Flush stages subcommand
      class Stages < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp flush stages
Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.new(cli_options: config, patterns: cli_arguments).flush_stages
        end
      end
    end
  end
end
