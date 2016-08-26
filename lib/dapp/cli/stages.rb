module Dapp
  class CLI
    # CLI stages subcommand
    class Stages < CLI
      SUBCOMMANDS = ['flush local', 'flush repo', 'cleanup local', 'cleanup repo', 'push'].freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp stages cleanup local
dapp stages cleanup repo
dapp stages flush local
dapp stages flush repo
dapp stages push

Options:
BANNER
    end
  end
end
