require 'mixlib/cli'

module Dapp
  # CLI
  class CLI
    include Mixlib::CLI

    class << self
      def parse_options(cli, argv)
        cli.parse_options(argv)
      rescue OptionParser::InvalidOption => e
        STDERR.puts "Error: #{e.message}"
        puts
        puts cli.opt_parser
        exit 1
      end
    end

    banner <<BANNER.freeze
Usage: dapp [options] sub-command [sub-command options]

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp build [options] [PATTERN ...]
dapp list [options] [PATTERN ...]

Options:
BANNER

    option :version,
           long: '--version',
           description: 'Show version',
           on: :tail,
           boolean: true,
           proc: proc { puts "dapp: #{Dapp::VERSION}" },
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
      opt_parser.version = Dapp::VERSION
    end

    SUBCOMMANDS = %w(build list).freeze

    def parse_subcommand(argv)
      if (index = argv.find_index { |v| SUBCOMMANDS.include? v })
        return [
          argv[0...index],
          argv[index],
          argv[index.next..-1]
        ]
      else
        return [
          argv,
          nil,
          []
        ]
      end
    end

    def run(argv = ARGV)
      argv, subcommand, subcommand_argv = parse_subcommand(argv)

      CLI.parse_options(self, argv)

      run_subcommand subcommand, subcommand_argv
    end

    def run_subcommand(subcommand, subcommand_argv)
      if subcommand
        self.class.const_get(subcommand.capitalize).new.run(subcommand_argv)
      else
        STDERR.puts 'Error: subcommand not passed'
        puts
        puts opt_parser
        exit 1
      end
    end
  end
end
