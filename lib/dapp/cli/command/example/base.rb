module Dapp
  class CLI
    module Command
      class Example < ::Dapp::CLI
        class Base < Base
          option :examples_repo,
                 long:        '--examples-repo GIT_REPO',
                 description: 'Git repository with examples (\'https://github.com/flant/dapp.git\' by default).',
                 default:     'https://github.com/flant/dapp.git'

          option :examples_branch,
                 long:        '--examples-branch GIT_BRANCH',
                 description: 'Specific git branch (\'master\' by default).',
                 default:     'master'

          option :examples_dir,
                 long:        '--examples-dir DIR',
                 description: 'Directory with examples (\'examples\' by default)',
                 default:     'examples'

          def run_method
            "example_#{super}"
          end
        end
      end
    end
  end
end
