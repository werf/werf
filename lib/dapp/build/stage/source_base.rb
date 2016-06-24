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

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              git_artifact.send(apply_method, image, self)
            end
            yield image if block_given?
          end
        end

        # FIXME move to patch_apply_command
        def layer_actual?(git_artifact)
          prev_commit = git_artifact.layer_commit(prev_source_stage)
          current_commit = git_artifact.layer_commit(self)
          prev_commit == current_commit and !git_artifact.any_changes?(prev_commit, current_commit)
        end

        protected

        def apply_method
          :layer_apply!
        end

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
