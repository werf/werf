module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Tag < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg tag [options] [DIMG] TAG
  
    DIMG                        Dapp image to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          tag = self.class.required_argument(self)
          ::Dapp::Dapp.new(options: cli_options(dimgs_patterns: cli_arguments)).public_send(class_to_lowercase, tag)
        end
      end
    end
  end
end
