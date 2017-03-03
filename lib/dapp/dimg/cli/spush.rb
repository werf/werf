module Dapp
  module Dimg
    module CLI
      # CLI spush subcommand
      class Spush < Push
        banner <<BANNER.freeze
Usage:

  dapp dimg spush [options] [DIMG] REPO

    DIMG                        Dapp image to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER
      end
    end
  end
end
