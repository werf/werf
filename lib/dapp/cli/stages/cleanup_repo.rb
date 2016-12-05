module Dapp
  class CLI
    class Stages
      # stages cleanup repo subcommand
      class CleanupRepo < CleanupLocal
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages cleanup repo [options] [DIMG ...] REPO

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
