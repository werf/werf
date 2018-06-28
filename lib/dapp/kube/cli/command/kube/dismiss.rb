module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class Dismiss < Base
      banner <<BANNER.freeze
Usage:

  dapp kube dismiss [options]

Options:
BANNER

      option :namespace,
             long: '--namespace NAME',
             default: nil

      option :context,
             long: '--context NAME',
             default: nil

      option :with_namespace,
             long: '--with-namespace',
             default: false

      option :kubernetes_timeout,
             long: '--kubernetes-timeout TIMEOUT',
             description: 'Kubernetes api-server tcp connection, read and write timeout (in seconds)',
             proc: ->(v) { v.to_i }

    end
  end
end
