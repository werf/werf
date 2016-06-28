module Dapp
  module Build
    module Stage
      class Source4 < SourceBase
        MAX_PATCH_SIZE = 1024*1024

        def initialize(application, relative_stage)
          @prev_stage = AppSetup.new(application, self)
          super
        end

        def next_source_stage
          next_stage
        end

        def signature
          change_commit_if_patch_to_big!
          super
        end

        private

        def change_commit_if_patch_to_big!
          application.git_artifact_list.each do |git_artifact|
            layer_commit_change(git_artifact) if git_artifact.patch_size(layer_commit(git_artifact), git_artifact.repo_latest_commit) > MAX_PATCH_SIZE
          end
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
