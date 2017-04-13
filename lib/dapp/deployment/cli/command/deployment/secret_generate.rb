module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class SecretGenerate < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment secret generate

Options:
BANNER
    end
  end
end
