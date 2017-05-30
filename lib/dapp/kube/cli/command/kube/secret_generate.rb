module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretGenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret generate

Options:
BANNER
    end
  end
end
