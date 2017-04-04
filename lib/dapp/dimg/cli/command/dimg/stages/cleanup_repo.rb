module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class CleanupRepo < CleanupLocal
          banner <<BANNER.freeze
Usage:

  dapp dimg stages cleanup repo [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
          def run_method
            :stages_cleanup_repo
          end

          def repo
            self.class.required_argument(self)
          end
        end
      end
    end
  end
end
