module Dapp
  # CliHelper
  module CliHelper
    def method_name
      self.class.to_s.split('::').last.downcase.to_s
    end

    # ClassMethods
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

      # rubocop:disable Metrics/MethodLength
      def parse_subcommand(cli, args)
        argv = args
        divided_subcommand = []
        subcommand_argv = []

        cmd_arr = args.dup
        loop do
          if cli.class::SUBCOMMANDS.include? cmd_arr.join(' ')
            argv = args[0...args.index(cmd_arr.first)]
            divided_subcommand = cmd_arr
            index = cmd_arr.one? ? args.index(cmd_arr.first).next : args.index(cmd_arr.last).next
            subcommand_argv = args[index..-1]
          elsif !cmd_arr.empty?
            cmd_arr.pop
            next
          end
          break
        end

        [argv, divided_subcommand, subcommand_argv]
      end
      # rubocop:enable Metrics/MethodLength

      def run_subcommand(cli, divided_subcommand, subcommand_argv)
        if !divided_subcommand.empty?
          cli.class.const_get(prepare_subcommand(divided_subcommand)).new.run(subcommand_argv)
        else
          STDERR.puts 'Error: subcommand not passed'
          puts
          puts cli.opt_parser
          exit 1
        end
      end

      def prepare_subcommand(divided_subcommand)
        Array(divided_subcommand).map(&:capitalize).join
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
