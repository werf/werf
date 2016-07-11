require 'mixlib/cli'

module Dapp
  class CLI
    class Smartpush < Push
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp smartpush [options] [PATTERN ...] REPOPREFIX

    PATTERN                     Applications to process [default: *].

Options:
BANNER

    end
  end
end
