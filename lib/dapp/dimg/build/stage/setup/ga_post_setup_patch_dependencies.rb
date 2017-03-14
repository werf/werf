module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class GAPostSetupPatchDependencies < GADependenciesBase
            include Mod::Group

            MAX_PATCH_SIZE = 1024 * 1024

            def initialize(dimg, next_stage)
              @prev_stage = Setup.new(dimg, self)
              super
            end

            def dependencies
              [(changes_size_since_g_a_pre_setup_patch / MAX_PATCH_SIZE).to_i]
            end

            private

            def changes_size_since_g_a_pre_setup_patch
              dimg.git_artifacts.map do |git_artifact|
                git_artifact.patch_size(prev_g_a_stage.layer_commit(git_artifact), git_artifact.latest_commit)
              end.reduce(0, :+)
            end
          end # GAPostSetupPatchDependencies
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
