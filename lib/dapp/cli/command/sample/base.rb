module Dapp
  class CLI
    module Command
      class Sample < ::Dapp::CLI
        class Base < Base
          option :samples_repo,
                 long:        '--samples-repo GIT_REPO',
                 description: 'Git repository with samples (\'https://github.com/flant/dapp.git\' by default)',
                 default:     'https://github.com/flant/dapp.git'

          option :samples_branch,
                 long:        '--samples-branch GIT_BRANCH',
                 description: 'Git branch (\'master\' by default)',
                 default:     'master'

          option :samples_dir,
                 long:        '--samples-dir DIR',
                 description: 'Directory with samples (\'samples\' by default)',
                 default:     'samples'

          def run_method
            "sample_#{super}"
          end
        end
      end
    end
  end
end
