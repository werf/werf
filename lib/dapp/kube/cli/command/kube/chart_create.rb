module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class ChartCreate < Base
      banner <<BANNER.freeze
Usage:

  dapp kube chart create [options]

Options:
BANNER

      option :force,
             long: '--force',
             short: '-f',
             description: 'Delete existing chart',
             default: false
    end
  end
end
