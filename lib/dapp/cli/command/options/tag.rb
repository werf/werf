module Dapp
  class CLI
    module Command
      module Options
        module Tag
          def self.extended(klass)
            klass.class_eval do
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
            end
          end
        end
      end
    end
  end
end
