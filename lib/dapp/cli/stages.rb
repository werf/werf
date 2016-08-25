module Dapp
  class CLI
    # CLI stages subcommand
    class Stages < CLI
      SUBCOMMANDS = ['flush local', 'flush repo', 'cleanup local', 'cleanup repo'].freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp stages cleanup local
dapp stages cleanup repo
dapp stages flush local
dapp stages flush repo

Options:
BANNER
    end
  end
end
