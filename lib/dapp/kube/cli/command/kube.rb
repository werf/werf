module Dapp::Kube::CLI
  module Command
    class Kube < ::Dapp::CLI
      SUBCOMMANDS = ['secret generate', 'secret key generate'].freeze

      banner <<BANNER.freeze
Usage: dapp kube sub-command [sub-command options]

Available subcommands: (for details, dapp kube SUB-COMMAND --help)

  dapp kube secret generate [options]
  dapp kube secret key generate [options]

Options:
BANNER
    end
  end
end
