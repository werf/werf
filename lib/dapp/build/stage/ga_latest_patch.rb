module Dapp
  module Build
    module Stage
      # GALatestPatch
      class GALatestPatch < GABase
        def initialize(dimg, next_stage)
          @prev_stage = AfterSetupArtifact.new(dimg, self)
          super
        end

        def prev_g_a_stage
          prev_stage.prev_stage # GAPostSetupPatch
        end

        def next_g_a_stage
          nil
        end

        def dependencies_stage
          nil
        end

        def dependencies
          [commit_list]
        end

        def layer_commit(git_artifact)
          commits[git_artifact.full_name] ||= git_artifact.latest_commit
        end

        def empty?
          dependencies_empty?
        end

        private

        def commit_list
          dimg.git_artifacts.map { |git_artifact| layer_commit(git_artifact) }
        end
      end # GALatestPatch
    end # Stage
  end # Build
end # Dapp
