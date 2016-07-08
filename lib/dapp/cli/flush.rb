require 'mixlib/cli'

module Dapp
  class CLI
    class Flush < CLI
      SUBCOMMANDS = %w(stage).freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp flush stage
BANNER

    end
  end
end
