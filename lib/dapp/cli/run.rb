module Dapp
  class CLI
    # CLI run subcommand
    class Run < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp run [options] [DIMG PATTERN] [DOCKER ARGS]

    DIMG PATTERN                Dapp image to process [default: *].
    DOCKER ARGS                 Docker run options and command separated by '--'

Options:
BANNER

      def read_cli_options(args)
        self.class.cli_wrapper(self) do
          args.each_with_index do |arg, i|
            next if arg == '--'
            next if (key = find_option(arg)).nil?
            cli_option = []
            cli_option << args.slice!(i)
            if key[:with_arg]
              raise OptionParser::InvalidOption if args.count < i + 1
              cli_option << args.slice!(i)
            end
            parse_options(cli_option)
            return read_cli_options(args)
          end
        end
      end

      def find_option(arg)
        expected_options.each { |hash| return hash if hash[:formats].any? { |f| f.start_with? arg } }
        nil
      end

      def expected_options
        @expected_options ||= options.values.map { |opt| { formats: [opt[:long], opt[:short]].compact, with_arg: !opt[:long].split.one? } }
      end

      def run(argv = ARGV)
        filtered_args = read_cli_options(argv)
        pattern = filtered_args.any? && !filtered_args.first.start_with?('-') ? [filtered_args.shift] : []
        index = filtered_args.index('--') || filtered_args.count
        docker_options = index.nonzero? ? filtered_args.slice(0..index - 1) : []
        command = filtered_args.slice(index + 1..-1) || []
        Project.new(cli_options: config, dimgs_patterns: pattern).run(docker_options, command)
      end
    end
  end
end
