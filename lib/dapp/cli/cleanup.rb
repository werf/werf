require 'mixlib/cli'

module Dapp
  class CLI
    # Cleanup subcommand
    class Cleanup < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp cleanup [options] [APPS PATTERN ...]

    APPS PATTERN                Applications to process [default: *].

Options:
BANNER
    end
  end
end
