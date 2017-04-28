module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class Apply < Base
      banner <<BANNER.freeze
Usage:

  dapp deploy apply [options] REPO IMAGE_VERSION

Options:
BANNER

      option :namespace,
             long: '--namespace NAME'

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        image_version = self.class.required_argument(self, 'image_version')
        repo = self.class.required_argument(self, 'repo')
        ::Dapp::Dapp.new(options: cli_options(dimgs_patterns: cli_arguments)).deployment_apply(repo, image_version)
      end
    end
  end
end
