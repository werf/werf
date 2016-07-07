require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    module Base
      include Mixlib::CLI

      option :log_quiet,
             short: '-q',
             long: '--quiet',
             description: 'Suppress logging',
             on: :tail,
             boolean: true,
             builder_opt: true

      option :log_verbose,
             long: '--verbose',
             description: 'Enable verbose output',
             on: :tail,
             boolean: true,
             builder_opt: true

      def self.composite_options(opt)
        @composite_options ||= {}
        @composite_options[opt] ||= []
      end

      def initialize
        self.class.options.merge!(Base.options)
        super()
      end

      def run(argv = ARGV)
        CLI.parse_options(self, argv)
        NotBuilder.new(cli_options: config, patterns: cli_arguments).public_send(method_name)
      end

      def method_name
        self.class.to_s.split('::').last.downcase.to_s
      end

      def self.included(base)
        base.include Mixlib::CLI
      end
    end
  end
end
