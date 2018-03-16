module Dapp
  class CLI
    module Command
      class Base < ::Dapp::CLI
        option :dir,
               long: '--dir PATH',
               description: 'Change to directory',
               on: :head

        option :build_dir,
               long: '--build-dir PATH',
               description: 'Directory where build cache stored (DIR/.dapp_build by default)'

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

        option :ignore_config_warning,
               long: '--ignore-config-sequential-processing-warnings',
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

        option :dev,
               long: '--dev',
               default: false,
               boolean: true

        def initialize
          self.class.options.merge!(Base.options)
          super()
        end

        def run(_argv = ARGV)
          raise
        end

        def cli_options(**kvargs)
          config.merge(dapp_command: run_method, **kvargs)
        end
      end
    end
  end
end
