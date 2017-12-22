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
        extend ::Dapp::CLI::Command::Options::Tag

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          run_dapp_command(nil, options: cli_options(dimgs_patterns: cli_arguments)) do |dapp|
            repo = if not cli_arguments[0].nil?
              self.class.required_argument(self, 'repo')
            else
              dapp.name
            end

            dapp.options[:repo] = repo

            dapp.public_send(run_method)
          end
        end
      end
    end
  end
end
