module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class MinikubeSetup < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment minikube setup

Options:
BANNER
    end
  end
end
