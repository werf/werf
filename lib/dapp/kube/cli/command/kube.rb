module Dapp::Kube::CLI
  module Command
    class Kube < ::Dapp::CLI
      SUBCOMMANDS = ['secret generate', 'secret key generate', 'secret regenerate', 'deploy', 'dismiss', 'secret extract', 'secret edit', 'minikube setup', 'chart create', 'render', 'lint'].freeze

      banner <<BANNER.freeze
Usage: dapp kube subcommand [subcommand options]

Available subcommands: (for details, dapp kube SUB-COMMAND --help)

  dapp kube deploy [options] [REPO]
  dapp kube dismiss [options]
  dapp kube secret key generate [options]
  dapp kube secret generate [FILE_PATH] [options]
  dapp kube secret extract [FILE_PATH] [options]
  dapp kube secret regenerate [SECRET_VALUES_FILE_PATH ...] [options]
  dapp kube secret edit [FILE_PATH] [options]
  dapp kube minikube setup
  dapp kube chart create [options]
  dapp kube render [options] [REPO]
  dapp kube lint [options] [REPO]

Options:
BANNER
    end
  end
end
