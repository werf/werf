module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
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
