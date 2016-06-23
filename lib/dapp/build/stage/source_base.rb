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

        protected

        def commit_list
          build.git_artifact_list.map { |git_artifact| git_artifact.layer_commit(self) }
        end

        def layers_actual?
          build.git_artifact_list.map { |git_artifact| git_artifact.layer_actual?(self) }.all?
        end
      end # SourceBase
    end # Stage
  end # Build
end # Dapp
