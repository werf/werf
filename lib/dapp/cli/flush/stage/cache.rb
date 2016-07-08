require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      class Cache < Stage
        SUBCOMMANDS = %w(cache).freeze

        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp flush stage cache
Options:
BANNER

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          NotBuilder.flush_stage_cache
        end
      end
    end
  end
end
