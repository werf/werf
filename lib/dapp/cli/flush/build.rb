require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      class Build < Flush
        SUBCOMMANDS = %w(cache).freeze

        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp flush build cache
BANNER

      end
    end
  end
end
