module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Stages < ::Dapp::CLI
        class Base < Base
          def run_dapp_command(run_command, options: {})
            super(run_command, options: options.merge(verbose: true))
          end
        end
      end
    end
  end
end
