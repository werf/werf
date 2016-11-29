module Dapp
  class CLI
    # CLI stages subcommand
    class Stages < CLI
      SUBCOMMANDS = ['flush local', 'flush repo', 'cleanup local', 'cleanup repo', 'push', 'pull'].freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp stages cleanup local [options] [DIMG ...] REPO
dapp stages cleanup repo [options] [DIMG ...] REPO
dapp stages flush local [options] [DIMG ...]
dapp stages flush repo [options] [DIMG ...] REPO
dapp stages push [options] [DIMG ...] REPO
dapp stages pull [options] [DIMG ...] REPO

Options:
BANNER
    end
  end
end
