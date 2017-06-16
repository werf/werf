module Dapp::Deployment::CLI
  module Command
    class Deployment < ::Dapp::CLI
      SUBCOMMANDS = ['apply', 'mrproper'].freeze

      banner <<BANNER.freeze
Usage: dapp deployment subcommand [subcommand options]

Available subcommands: (for details, dapp deployment SUB-COMMAND --help)

  dapp deployment apply [options] [APP ...] REPO
  dapp deployment mrproper [options]

Options:
BANNER
    end
  end
end
