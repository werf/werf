require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      # Flush build cache subcommand
      class BuildCache < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp flush build cache
Options:
BANNER
        option :build_cache_dir,
               long: '--build-cache-dir PATH',
               description: 'Build cache directory'

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          Controller.new(cli_options: config, patterns: cli_arguments).flush_build_cache
        end
      end
    end
  end
end
