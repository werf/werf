module Dapp
  class CLI
    # CLI spush subcommand
    class Spush < Push
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp spush [options] [DIMGS PATTERN ...] REPO

    DIMGS PATTERN               Dapp images to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER
    end
  end
end
