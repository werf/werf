module Dapp
  module Helper
    module Cli
      def parse_options(cli, argv)
        cli_wrapper(cli) do
          cli.parse_options(argv)
        end
      end

      def cli_wrapper(cli)
        yield
      rescue OptionParser::MissingArgument, OptionParser::InvalidOption, OptionParser::InvalidArgument, OptionParser::AmbiguousOption => e
        print_error_with_help_and_die!(cli, e.message)
      end

      def required_argument(cli, argument)
        unless (arg = cli.cli_arguments.pop)
          print_error_with_help_and_die!(cli, "required argument `#{argument.upcase}`")
        end
        arg
      end

      def print_error_with_help_and_die!(cli, error_message)
        STDERR.puts "Error: #{error_message}"
        puts
        puts cli.opt_parser
        exit 1
      end

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

      def run_subcommand(cli, divided_subcommand, subcommand_argv)
        if !divided_subcommand.empty?
          cli.class.const_get(prepare_subcommand(divided_subcommand)).new.run(subcommand_argv)
        else
          puts cli.opt_parser
          exit 1
        end
      end

      def prepare_subcommand(divided_subcommand)
        Array(divided_subcommand)
          .map { |c| c.split(/[-_]/) }
          .flatten
          .map(&:capitalize)
          .join
      end

      def composite_options(opt)
        @composite_options ||= {}
        @composite_options[opt] ||= []
      end

      def in_validate!(v, list)
        raise OptionParser::InvalidArgument, "`#{v}` is not included in the list [#{list_msg_format(list)}]" unless list.include?(v)
      end

      def list_msg_format(list)
        list.map { |s| "'#{s}'"}.join(', ')
      end
    end
  end # Helper
end # Dapp
