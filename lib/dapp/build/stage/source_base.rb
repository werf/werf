module Dapp
  module Build
    module Stage
      class SourceBase < Base
        attr_accessor :prev_source_stage, :next_source_stage

        def prev_source_stage
          prev_stage.prev_stage
        end

        def next_source_stage
          next_stage.next_stage
        end

        def image(&blk)
          blk = ->(image) do
            build.git_artifact_list.each do |git_artifact|
              git_artifact.layer_apply!(image, self)
            end
          end unless block_given?
          super(&blk)
        end

        def layer_actual?(git_artifact)
          prev_commit = git_artifact.layer_commit(prev_source_stage)
          current_commit = git_artifact.layer_commit(self)
          prev_commit == current_commit and !git_artifact.any_changes?(prev_commit, current_commit)
        end

        protected

        def commit_list
          build.git_artifact_list.map { |git_artifact| git_artifact.layer_commit(self) }
        end

        def layers_actual?
          build.git_artifact_list.map { |git_artifact| layer_actual?(git_artifact) }.all?
        end
      end # SourceBase
    end # Stage
  end # Build
end # Dapp
