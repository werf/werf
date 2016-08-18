require 'mixlib/cli'

module Dapp
  class CLI
    # Base of CLI subcommands
    class Base < CLI
      option :dir,
             long: '--dir PATH',
             description: 'Change to directory',
             on: :head

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Directory where build cache stored (DIR/.dapps-build by default)'

      option :log_quiet,
             short: '-q',
             long: '--quiet',
             description: 'Suppress logging',
             default: false,
             boolean: true

      option :log_verbose,
             long: '--verbose',
             description: 'Enable verbose output',
             default: false,
             boolean: true

      option :log_time,
             long: '--time',
             description: 'Enable output with time',
             default: false,
             boolean: true

      option :log_color,
             long: '--color MODE',
             description: 'Display output in color on the terminal',
             in: %w(auto on off),
             default: 'auto'

      option :dry_run,
             long: '--dry-run',
             default: false,
             boolean: true

      def initialize
        self.class.options.merge!(Base.options)
        super()
      end

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        Controller.new(cli_options: config, patterns: cli_arguments).public_send(class_to_lowercase)
      end
    end
  end
end
