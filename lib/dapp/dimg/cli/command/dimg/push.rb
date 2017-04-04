module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Push < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg push [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
        option :lock_timeout,
               long: '--lock-timeout TIMEOUT',
               description: 'Redefine resource locking timeout (in seconds)',
               proc: ->(v) { v.to_i }

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

        option :with_stages,
               long: '--with-stages',
               boolean: true

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self)
          ::Dapp::Dapp.new(options: cli_options(dimgs_patterns: cli_arguments)).public_send(class_to_lowercase, repo)
        end
      end
    end
  end
end
