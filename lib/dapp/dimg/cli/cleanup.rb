module Dapp
  module Dimg
    module CLI
      class Cleanup < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg cleanup [options] [DIMG ...]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER

        option :lock_timeout,
               long: '--lock-timeout TIMEOUT',
               description: 'Redefine resource locking timeout (in seconds)',
               proc: ->(v) { v.to_i }
      end
    end
  end
end
