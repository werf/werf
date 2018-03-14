module Dapp
  module Dimg
    module Build
      module Stage
        class GALatestPatch < GABase
          def initialize(dimg, next_stage)
            @prev_stage = AfterSetupArtifact.new(dimg, self)
            super
          end

          def renew
            dependencies_discard
            super
          end

          def dependencies
            @dependencies ||= [commit_list, git_artifacts_dev_patch_hashes]
          end

          def empty?
            dimg.git_artifacts.empty? || dependencies_empty?
          end

          def layer_commit(git_artifact)
            commits[git_artifact] ||= git_artifact.latest_commit
          end

          private

          def commit_list
            dimg.git_artifacts
              .select { |ga| ga.repo.commit_exists?(prev_stage.layer_commit(ga)) && ga.patch_any_changes?(self) }
              .map(&method(:layer_commit))
          end

          def git_artifacts_dev_patch_hashes
            dimg.git_artifacts.map(&:dev_patch_hash)
          end
        end # GALatestPatch
      end # Stage
    end # Build
  end # Dimg
end # Dapp
