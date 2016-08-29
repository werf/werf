module Dapp
  class CLI
    # CLI stages subcommand
    class Stages < CLI
      SUBCOMMANDS = ['flush local', 'flush repo', 'cleanup local', 'cleanup repo', 'push', 'pull'].freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp stages cleanup local
dapp stages cleanup repo
dapp stages flush local
dapp stages flush repo
dapp stages push
dapp stages pull

Options:
BANNER
    end
  end
end
