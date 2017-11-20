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

            def renew
              dependencies_discard
              super
            end

            def dependencies
              @dependencies ||= [(changes_size_since_g_a_pre_setup_patch / MAX_PATCH_SIZE).to_i]
            end

            private

            def changes_size_since_g_a_pre_setup_patch
              dimg.git_artifacts.map do |git_artifact|
                if git_artifact.repo.commit_exists? prev_stage.layer_commit(git_artifact)
                  git_artifact.patch_size(prev_stage.layer_commit(git_artifact), git_artifact.latest_commit)
                else
                  0
                end
              end.reduce(0, :+)
            end
          end # GAPostSetupPatchDependencies
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
