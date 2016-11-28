module Dapp
  class CLI
    # CLI spush subcommand
    class Spush < Push
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp spush [options] [DIMG PATTERN ...] REPO

    DIMG PATTERN                Dapp image to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER
    end
  end
end
