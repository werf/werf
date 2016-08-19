module Dapp
  module Build
    module Stage
      # GAPostSetupPatchDependencies
      class GAPostSetupPatchDependencies < GADependenciesBase
        MAX_PATCH_SIZE = 1024 * 1024

        def initialize(application, next_stage)
          @prev_stage = ChefCookbooks.new(application, self)
          super
        end

        def dependencies
          [(changes_size_since_g_a_pre_setup_patch / MAX_PATCH_SIZE).to_i]
        end

        private

        def changes_size_since_g_a_pre_setup_patch
          application.git_artifacts.map do |git_artifact|
            git_artifact.patch_size(prev_stage.prev_stage.prev_stage.layer_commit(git_artifact), git_artifact.latest_commit)
          end.reduce(0, :+)
        end
      end # GAPostSetupPatchDependencies
    end # Stage
  end # Build
end # Dapp
