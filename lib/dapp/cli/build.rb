require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Build
      include Base

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp build [options] [PATTERN ...]

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

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Build directory'

      option :build_cache_dir,
             long: '--build-cache-dir PATH',
             description: 'Build cache directory'

      def run(argv = ARGV)
        CLI.parse_options(self, argv)
        NotBuilder.new(cli_options: config, patterns: cli_arguments).build
      end
    end
  end
end
