module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class Mrproper < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment mrproper [options]

Options:
BANNER
    end
  end
end
