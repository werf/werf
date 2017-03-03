module Dapp
  module Dimg
    module CLI
      # CLI list subcommand
      class List < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg list [options] [DIMG ...]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
      end
    end
  end
end
