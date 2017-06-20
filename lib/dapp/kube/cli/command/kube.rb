module Dapp::Kube::CLI
  module Command
    class Kube < ::Dapp::CLI
      SUBCOMMANDS = ['secret generate', 'secret key generate', 'deploy', 'dismiss', 'secret extract', 'minikube setup'].freeze

      banner <<BANNER.freeze
Usage: dapp kube subcommand [subcommand options]

Available subcommands: (for details, dapp kube SUB-COMMAND --help)

  dapp kube deploy [options] REPO
  dapp kube dismiss [options]
  dapp kube secret key generate [options]
  dapp kube secret generate [FILE_PATH] [options]
  dapp kube secret extract [FILE_PATH] [options]
  dapp kube minikube setup

Options:
BANNER
    end
  end
end
