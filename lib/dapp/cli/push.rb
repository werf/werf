require 'mixlib/cli'

module Dapp
  class CLI
    # CLI push subcommand
    class Push < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp push [options] [PATTERN ...] REPO

    PATTERN                     Applications to process [default: *].
    REPO                        Pushed image name.
Options:
BANNER

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Directory where build cache stored (DIR/.dapps-build by default)'

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from'

      option :tag,
             long: '--tag TAG',
             description: 'Add tag (can be used one or more times)',
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

      option :force,
             long: '--force',
             description: 'Override existing image',
             default: false,
             boolean: true

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        repo = self.class.required_argument(self)
        Controller.new(cli_options: config, patterns: cli_arguments).public_send(class_to_lowercase, repo)
      end
    end
  end
end
