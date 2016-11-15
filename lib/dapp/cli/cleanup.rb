module Dapp
  class CLI
    # Cleanup subcommand
    class Cleanup < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp cleanup [options] [DIMGS PATTERN ...]

    DIMGS PATTERN               Dapp images to process [default: *].

Options:
BANNER

      option :lock_timeout,
             long: '--lock-timeout TIMEOUT',
             description: 'Redefine resource locking timeout (in seconds)',
             proc: ->(v) { v.to_i }
    end
  end
end
