module Dapp
  module Dimg
    module CLI
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
