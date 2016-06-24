module Dapp
  module Build
    module Stage
      class Source4 < SourceBase
        MAX_PATCH_SIZE = 1024*1024

        def initialize(build, relative_stage)
          @prev_stage = AppSetup.new(build, self)
          super
        end

        def next_source_stage
          next_stage
        end

        def signature
          if latest_patch_to_big?
            hashsum prev_stage.signature
          else
            hashsum [dependencies_checksum, *commit_list]
          end
        end

        private

        def latest_patch_to_big?
          build.git_artifact_list.all? do |git_artifact|
            git_artifact.patch_size(layer_commit(git_artifact), git_artifact.repo_latest_commit) < MAX_PATCH_SIZE
          end
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
