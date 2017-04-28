module Dapp::Deployment::CLI
  module Command
    class Deployment < ::Dapp::CLI
      SUBCOMMANDS = ['apply', 'mrproper', 'secret generate', 'secret key generate', 'minikube setup'].freeze

      banner <<BANNER.freeze
Usage: dapp deployment sub-command [sub-command options]

Available subcommands: (for details, dapp deployment SUB-COMMAND --help)

  dapp deployment apply [options] [APP ...] REPO
  dapp deployment secret key generate
  dapp deployment secret generate
  dapp deployment minikube setup
  dapp deployment mrproper [options]

Options:
BANNER
    end
  end
end
