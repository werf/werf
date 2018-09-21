module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Tag < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg tag [options] [DIMG ...] REPO
  
    DIMG                        Dapp image to process [default: *].

Options:
BANNER
        extend ::Dapp::CLI::Options::Tag

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self, 'repo')
          run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments, repo: repo, ignore_try_host_docker_login: true))
        end
      end
    end
  end
end
