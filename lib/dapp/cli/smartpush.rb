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

      def run(argv = ARGV)
        CLI.parse_options(self, argv)
        repo_prefix = CLI.required_argument(self)
        NotBuilder.new(cli_options: config, patterns: cli_arguments).smartpush(repo_prefix)
      end
    end
  end
end
