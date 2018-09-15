module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Tag < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg tag [options] [DIMG ...] [REPO]
  
    DIMG                        Dapp image to process [default: *].

Options:
BANNER
        extend ::Dapp::CLI::Options::Tag

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          options = cli_options(dimgs_patterns: cli_arguments)
          options[:repo] = if not cli_arguments[0].nil?
                             self.class.required_argument(self, 'repo')
                           else
                             dapp.name
                           end
          run_dapp_command(run_method, options: options)
        end
      end
    end
  end
end
