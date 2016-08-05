require 'mixlib/cli'

module Dapp
  class CLI
    class Metadata
      # Metadata flush subcommand
      class Flush < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp metadata flush
Options:
BANNER
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.new(cli_options: config, patterns: cli_arguments).metadata_flush
        end
      end
    end
  end
end
