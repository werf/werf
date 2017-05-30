module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class SecretKeyGenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp kube secret key generate

Options:
BANNER
    end
  end
end
