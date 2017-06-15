module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class MinikubeSetup < Base
      banner <<BANNER.freeze
Usage:

  dapp kube minikube setup

Options:
BANNER
    end
  end
end
