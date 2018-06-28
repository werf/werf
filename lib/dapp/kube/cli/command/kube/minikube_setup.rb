module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class MinikubeSetup < Base
      banner <<BANNER.freeze
Usage:

  dapp kube minikube setup

Options:
BANNER

      option :kubernetes_timeout,
        long: '--kubernetes-timeout TIMEOUT',
        description: 'Kubernetes api-server tcp connection, read and write timeout (in seconds)',
        proc: ->(v) { v.to_i }

    end
  end
end
