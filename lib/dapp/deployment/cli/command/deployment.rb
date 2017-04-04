module Dapp::Deployment::CLI
  module Command
    class Deployment < ::Dapp::CLI
      SUBCOMMANDS = ['apply', 'mrproper'].freeze

      banner <<BANNER.freeze
Usage: dapp deployment sub-command [sub-command options]

Available subcommands: (for details, dapp deployment SUB-COMMAND --help)

  dapp deployment apply [options]
  dapp deployment mrproper [options]

Options:
BANNER
    end
  end
end
