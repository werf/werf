module Dapp
  module Build
    module Stage
      # GALatestPatch
      class GALatestPatch < GABase
        def initialize(application, next_stage)
          @prev_stage = SetupGroup::GAPostSetupPatch.new(application, self)
          super
        end

        def prev_g_a_stage
          prev_stage
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
          commits[git_artifact] ||= begin
            git_artifact.latest_commit
          end
        end

        def empty?
          dependencies_empty?
        end

        private

        def commit_list
          application.git_artifacts.map { |git_artifact| layer_commit(git_artifact) }
        end
      end # GALatestPatch
    end # Stage
  end # Build
end # Dapp
