module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class List < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg list [options] [DIMG ...]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
        def log_running_time
          false
        end
      end
    end
  end
end
