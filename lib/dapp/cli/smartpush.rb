require 'mixlib/cli'

module Dapp
  class CLI
    # CLI smartpush subcommand
    class Smartpush < Push
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp smartpush [options] [PATTERN ...] REPOPREFIX

    PATTERN                     Applications to process [default: *].
    REPOPREFIX                  Pushed image name prefix.
Options:
BANNER
    end
  end
end
