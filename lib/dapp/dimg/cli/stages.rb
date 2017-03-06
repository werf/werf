module Dapp
  module Dimg
    module CLI
      class Stages < ::Dapp::CLI
        SUBCOMMANDS = ['flush local', 'flush repo', 'cleanup local', 'cleanup repo', 'push', 'pull'].freeze

        banner <<BANNER.freeze
Available subcommands: (for details, dapp dimg stages SUB-COMMAND --help)

dapp dimg stages cleanup local [options] [DIMG ...] [REPO]
dapp dimg stages cleanup repo [options] [DIMG ...] REPO
dapp dimg stages flush local [options] [DIMG ...]
dapp dimg stages flush repo [options] [DIMG ...] REPO
dapp dimg stages push [options] [DIMG ...] REPO
dapp dimg stages pull [options] [DIMG ...] REPO

Options:
BANNER
      end
    end
  end
end
