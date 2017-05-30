module Dapp::Deployment::CLI::Command
  class Deployment < ::Dapp::CLI
    class Apply < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment apply [options] [APP ...] REPO

Options:
BANNER

      option :namespace,
             long: '--namespace NAME'

      option :image_version,
             long: '--image-version IMAGE_VERSION',
             default: 'latest'

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        repo = self.class.required_argument(self, 'repo')
        ::Dapp::Dapp.new(options: cli_options(apps_patterns: cli_arguments, repo: repo)).deployment_apply
      end
    end
  end
end
