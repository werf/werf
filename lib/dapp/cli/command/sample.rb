module Dapp
  class CLI
    module Command
      class Sample < ::Dapp::CLI
        SUBCOMMANDS = %w(list create).freeze

        banner <<BANNER.freeze
Usage: dapp sample [options] subcommand [subcommand options]

Available subcommands: (for details, dapp sample SUB-COMMAND --help)

  dapp sample list [options]
  dapp sample create SAMPLE_NAME [options] 

Options:
BANNER
      end
    end
  end
end
