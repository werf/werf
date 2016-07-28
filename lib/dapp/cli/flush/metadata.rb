require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      # Flush metadata subcommand
      class Metadata < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp flush metadata
Options:
BANNER
        option :metadata_dir,
               long: '--metadata-dir PATH',
               description: 'Metadata directory'

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.new(cli_options: config, patterns: cli_arguments).flush_metadata
        end
      end
    end
  end
end
