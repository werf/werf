module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class Lint < Base
      banner <<BANNER.freeze
Usage:

  dapp kube lint [options] [REPO]

Options:
BANNER

      option :namespace,
             long: '--namespace NAME',
             default: nil

      option :tag,
             long: '--tag TAG',
             description: 'Custom tag',
             default: [],
             proc: proc { |v| composite_options(:tags) << v }

      option :tag_branch,
             long: '--tag-branch',
             description: 'Tag by git branch',
             boolean: true

      option :tag_build_id,
             long: '--tag-build-id',
             description: 'Tag by CI build id',
             boolean: true

      option :tag_ci,
             long: '--tag-ci',
             description: 'Tag by CI branch and tag',
             boolean: true

      option :tag_commit,
             long: '--tag-commit',
             description: 'Tag by git commit',
             boolean: true

      option :helm_set_options,
             long: '--set STRING_ARRAY',
             default: [],
             proc: proc { |v| composite_options(:helm_set) << v }

      option :helm_values_options,
             long: '--values FILE_PATH',
             default: [],
             proc: proc { |v| composite_options(:helm_values) << v }

      option :helm_secret_values_options,
             long: '--secret-values FILE_PATH',
             default: [],
             proc: proc { |v| composite_options(:helm_secret_values) << v }

      def run(argv = ARGV)
        self.class.parse_options(self, argv)

        run_dapp_command(nil, options: cli_options, log_running_time: false) do |dapp|
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
