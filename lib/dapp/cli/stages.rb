module Dapp
  class CLI
    # CLI stages subcommand
    class Stages < CLI
      SUBCOMMANDS = ['flush local', 'flush repo', 'cleanup local', 'cleanup repo', 'push', 'pull'].freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp stages cleanup local [options] [DIMGS PATTERN ...] REPO
dapp stages cleanup repo [options] [DIMGS PATTERN ...] REPO
dapp stages flush local [options] [DIMGS PATTERN ...]
dapp stages flush repo [options] [DIMGS PATTERN ...] REPO
dapp stages push [options] [DIMGS PATTERN ...] REPO
dapp stages pull [options] [DIMGS PATTERN ...] REPO

Options:
BANNER
    end
  end
end
