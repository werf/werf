module Dapp
  class CLI
    # CLI tag subcommand
    class Tag < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp tag [options] [DIMG PATTERN ...] TAG

    DIMG PATTERN                Dapp image to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        tag = self.class.required_argument(self)
        Project.new(cli_options: config, dimgs_patterns: cli_arguments).public_send(class_to_lowercase, tag)
      end
    end
  end
end
