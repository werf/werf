require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      # Flush stages subcommand
      class Stages < Flush
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp flush stages
Options:
BANNER

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.flush_stages
        end
      end
    end
  end
end
