require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      class Build
        class Cache < Base
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
            NotBuilder.new(cli_options: config, patterns: cli_arguments).flush_build_cache
          end
        end
      end
    end
  end
end
