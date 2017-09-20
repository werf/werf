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

        option :quiet,
               short: '-q',
               long: '--quiet',
               description: 'Suppress logging',
               default: false,
               boolean: true

        option :verbose,
               long: '--verbose',
               description: 'Enable verbose output',
               default: false,
               boolean: true

        option :time,
               long: '--time',
               description: 'Enable output with time',
               default: false,
               boolean: true

        option :ignore_config_warning,
               long: '--ignore-config-sequential-processing-warnings',
               default: false,
               boolean: true

        option :color,
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

        def run_dapp_command(run_method, *args)
          dapp = ::Dapp::Dapp.new(*args)
          begin
            if block_given?
              yield dapp
            elsif !run_method.nil?
              dapp.public_send(run_method)
            end
          end
        end

        def run(_argv = ARGV)
          raise
        end

        def cli_options(**kvargs)
          config.merge(**kvargs)
        end
      end
    end
  end
end
