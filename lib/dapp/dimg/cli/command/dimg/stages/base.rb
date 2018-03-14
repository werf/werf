module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class Base < Base
          def run_dapp_command(run_command, options: {}, **extra_options)
            super(run_command, options: options.merge(verbose: true), **extra_options)
          end
        end
      end
    end
  end
end
