module Dapp
  module Dimg
    module CLI
      class Base < ::Dapp::CLI::Base
        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          ::Dapp::Dapp.new(cli_options: config, dimgs_patterns: cli_arguments).public_send(run_method)
        end

        def run_method
          class_to_lowercase
        end
      end
    end
  end
end
