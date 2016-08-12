require 'mixlib/cli'

module Dapp
  class CLI
    # CLI push subcommand
    class Push < Spush
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp push [options] [APP PATTERN] REPO

    APP PATTERN                 Applications to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER
    end
  end
end
