module Dapp
  class CLI
    include Mixlib::CLI

    extend Helper::Cli
    include Helper::Trivia

    SUBCOMMANDS = ['dimg', 'deployment'].freeze

    banner <<BANNER.freeze
Usage: dapp sub-command [sub-command options]

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp dimg
dapp deployment

Options:
BANNER

    option :version,
           long: '--version',
           description: 'Show version',
           on: :tail,
           boolean: true,
           proc: proc { puts "dapp: #{::Dapp::VERSION}" },
           exit: 0

    option :help,
           short: '-h',
           long: '--help',
           description: 'Show this message',
           on: :tail,
           boolean: true,
           show_options: true,
           exit: 0

    def initialize(*args)
      super(*args)

      opt_parser.program_name = 'dapp'
      opt_parser.version = ::Dapp::VERSION
    end

    def run(argv = ARGV)
      argv, subcommand, subcommand_argv = self.class.parse_subcommand(self, argv)
      self.class.parse_options(self, argv)
      self.class.run_subcommand self, subcommand, subcommand_argv
    end
  end
end
