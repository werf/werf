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
             required: true

      option :image_version,
             long: '--image-version IMAGE_VERSION',
             default: 'latest'

      option :tmp_dir_prefix,
             long: '--tmp-dir-prefix PREFIX',
             description: 'Tmp directory prefix (/tmp by default). Used for build process service directories.'

      option :helm_set_options,
             long: '--set STRING_ARRAY',
             default: [],
             proc: proc { |v| composite_options(:helm_set) << v }

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        repo = self.class.required_argument(self, 'repo')
        ::Dapp::Dapp.new(options: cli_options(apps_patterns: cli_arguments, repo: repo)).public_send(run_method)
      end
    end
  end
end
