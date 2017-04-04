module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class BuildContext < ::Dapp::CLI
        class Import < Export
          banner <<BANNER.freeze
Usage:

  dapp dimg build-context import [options]

Options:
BANNER
        end
      end
    end
  end
end
