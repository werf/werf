require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Show
      include Base

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp show [options] [PATTERN ...]

    PATTERN                     Applications to process [default: *].

Options:
BANNER

      option :help,
             short: '-h',
             long: '--help',
             description: 'Show this message',
             on: :tail,
             boolean: true,
             show_options: true,
             exit: 0

      option :dir,
             long: '--dir PATH',
             description: 'Change to directory',
             on: :head
    end
  end
end
