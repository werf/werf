require 'mixlib/cli'

module Dapp
  class CLI
    # CLI flush subcommand
    class Flush < CLI
      SUBCOMMANDS = %w(metadata stages).freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp flush metadata
dapp flush stages
BANNER
    end
  end
end
