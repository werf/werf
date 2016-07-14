require 'mixlib/cli'

module Dapp
  class CLI
    class Flush < CLI
      SUBCOMMANDS = ['stage cache', 'build cache'].freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp flush stage cache
dapp flush build cache
BANNER

    end
  end
end
