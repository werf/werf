module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class BuildContext < ::Dapp::CLI
        SUBCOMMANDS = %w(import export).freeze

        banner <<BANNER.freeze
Available subcommands: (for details, dapp dimg build-context SUB-COMMAND --help)

  dapp dimg build-context export [options] [DIMG ...]
  dapp dimg build-context import [options]

Options:
BANNER
      end
    end
  end
end
