module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class BuildContext < ::Dapp::CLI
        class Export < Base
          banner <<BANNER.freeze
Usage:

  dapp dimg build-context export [options] [DIMG ...]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          option :build_context_directory,
                 long: '--build-context-directory DIR_PATH',
                 description: 'Path to the directory with context'

          def run_method
            :"build_context_#{class_to_lowercase}"
          end
        end
      end
    end
  end
end
