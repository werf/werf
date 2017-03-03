module Dapp
  module Dimg
    module CLI
      class Mrproper < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg mrprooper [options]

Options:
BANNER
        option :proper_all,
               long: '--all',
               boolean: true

        option :proper_cache_version,
               long: '--improper-cache-version-stages',
               boolean: true

        option :proper_dev_mode_cache,
               long: '--improper-dev-mode-cache',
               boolean: true
      end
    end
  end
end
