require 'mixlib/cli'

module Dapp
  class CLI
    class Flush
      class Stage
        class Cache < Stage
          banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp flush stage cache
Options:
BANNER

          def run(argv = ARGV)
            self.class.parse_options(self, argv)
            Controller.flush_stage_cache
          end
        end
      end
    end
  end
end
