module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class Deploy < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment deploy [options] REPO

Options:
BANNER
      option :namespace,
             long: '--namespace NAME',
             default: nil

      option :image_version,
             long: '--image-version IMAGE_VERSION',
             default: 'latest'

      option :tmp_dir_prefix,
             long: '--tmp-dir-prefix PREFIX',
             description: 'Tmp directory prefix (/tmp by default). Used for build process service directories.'

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        repo = self.class.required_argument(self, 'repo')
        ::Dapp::Dapp.new(options: cli_options(repo: repo)).public_send(run_method)
      end
    end
  end
end
