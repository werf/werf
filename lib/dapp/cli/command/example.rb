module Dapp
  class CLI
    module Command
      class Example < ::Dapp::CLI
        SUBCOMMANDS = %w(list create).freeze

        banner <<BANNER.freeze
Usage: dapp example [options] subcommand [subcommand options]

Available subcommands: (for details, dapp example SUB-COMMAND --help)

  dapp example list [options]
  dapp example create EXAMPLE_NAME [options] 

Options:
BANNER
      end
    end
  end
end
