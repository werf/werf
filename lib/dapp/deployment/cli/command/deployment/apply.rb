module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class Apply < Base
      banner <<BANNER.freeze
Usage:

  dapp deploy apply [options]

Options:
BANNER

      option :namespace,
             long: '--namespace NAME'
    end
  end
end
