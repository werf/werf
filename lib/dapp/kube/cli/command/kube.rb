module Dapp::Kube::CLI
  module Command
    class Kube < ::Dapp::CLI
      SUBCOMMANDS = ['secret generate', 'secret key generate', 'secret file encrypt', 'deploy', 'dismiss', 'minikube setup'].freeze

      banner <<BANNER.freeze
Usage: dapp kube subcommand [subcommand options]

Available subcommands: (for details, dapp kube SUB-COMMAND --help)

  dapp kube deploy [options] REPO
  dapp kube dismiss [options]
  dapp kube secret generate [options]
  dapp kube secret key generate [options]
  dapp kube secret file encrypt FILE_PATH [options]
  dapp kube minikube setup

Options:
BANNER
    end
  end
end
