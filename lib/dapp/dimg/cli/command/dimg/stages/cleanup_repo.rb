module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class CleanupRepo < CleanupLocal
          banner <<BANNER.freeze
Usage:

  dapp dimg stages cleanup repo [options] REPO

Options:
BANNER
          def run_method
            :stages_cleanup_repo
          end

          def repo
            self.class.required_argument(self, 'repo')
          end
        end
      end
    end
  end
end
