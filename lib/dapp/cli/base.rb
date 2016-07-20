require 'mixlib/cli'

module Dapp
  class CLI
    # Base of CLI subcommands
    class Base < CLI
      option :dir,
             long: '--dir PATH',
             description: 'Change to directory',
             on: :head

      option :log_quiet,
             short: '-q',
             long: '--quiet',
             description: 'Suppress logging',
             on: :tail,
             default: false,
             boolean: true

      option :log_verbose,
             long: '--verbose',
             description: 'Enable verbose output',
             on: :tail,
             default: false,
             boolean: true

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
        Controller.new(cli_options: config, patterns: cli_arguments).public_send(method_name)
      end
    end
  end
end
