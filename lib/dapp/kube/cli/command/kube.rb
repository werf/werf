module Dapp::Kube::CLI
  module Command
    class Kube < ::Dapp::CLI
      SUBCOMMANDS = [].freeze

      banner <<BANNER.freeze
Usage: dapp kube sub-command [sub-command options]

Available subcommands: (for details, dapp kube SUB-COMMAND --help)

Options:
BANNER
    end
  end
end
