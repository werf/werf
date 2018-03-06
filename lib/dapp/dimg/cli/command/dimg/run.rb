module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Run < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg run [options] [DIMG] [DOCKER ARGS]

    DIMG                        Dapp image to process [default: *].
    DOCKER ARGS                 Docker run options and command separated by '--'

Options:
BANNER
        option :stage,
               long:        '--stage STAGE',
               description: "Run one of the following stages (#{list_msg_format(DIMG_STAGES)})",
               proc:        STAGE_PROC.call(DIMG_STAGES)

        option :ssh_key,
               long: '--ssh-key SSH_KEY',
               description: ['Enable only specified ssh keys ',
                             '(use system ssh-agent by default)'].join,
               default: nil,
               proc: ->(v) { composite_options(:ssh_key) << v }

        def read_options(args)
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
              return read_options(args)
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
          filtered_args = read_options(argv)
          patterns = filtered_args.any? && !filtered_args.first.start_with?('-') ? [filtered_args.shift] : []
          index = filtered_args.index('--') || filtered_args.count
          docker_options = index.nonzero? ? filtered_args.slice(0..index - 1) : []
          command = filtered_args.slice(index + 1..-1) || []

          if docker_options.empty? && command.empty?
            docker_options = %w(-ti --rm)
            command = %w(/bin/bash)
          end

          stage_name = config.delete(:stage)

          run_dapp_command(nil, options: cli_options(dimgs_patterns: patterns), log_running_time: false) do |dapp|
            dapp.run(stage_name, docker_options, command)
          end
        end
      end
    end
  end
end
