module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class SecretKeyGenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment secret key generate

Options:
BANNER
    end
  end
end
