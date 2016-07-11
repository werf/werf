module Dapp
  module CliHelper
    def method_name
      self.class.to_s.split('::').last.downcase.to_s
    end

    module ClassMethods
      def parse_options(cli, argv)
        cli.parse_options(argv)
      rescue OptionParser::InvalidOption => e
        STDERR.puts "Error: #{e.message}"
        puts
        puts cli.opt_parser
        exit 1
      end

      def required_argument(cli)
        unless (arg = cli.cli_arguments.pop)
          puts
          puts cli.opt_parser
          exit 1
        end
        arg
      end

      def parse_subcommand(cli, argv)
        if (index = argv.find_index { |v| cli.class::SUBCOMMANDS.include? v })
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

      def run_subcommand(cli, subcommand, subcommand_argv)
        if subcommand
          cli.class.const_get(subcommand.capitalize).new.run(subcommand_argv)
        else
          STDERR.puts 'Error: subcommand not passed'
          puts
          puts cli.opt_parser
          exit 1
        end
      end

      def composite_options(opt)
        @composite_options ||= {}
        @composite_options[opt] ||= []
      end
    end

    def self.included(base)
      base.extend(ClassMethods)
    end
  end # CommonHelper
end # Dapp
