module Dapp
  module Dimg
    module Build
      module Stage
        class GALatestPatch < GABase
          def initialize(dimg, next_stage)
            @prev_stage = AfterSetupArtifact.new(dimg, self)
            super
          end

          def dependencies_stage
            nil
          end

          def dependencies
            [].tap do |dependencies|
              dependencies << commit_list
              dependencies << dimg.local_git_artifacts.map { |git_artifact| git_artifact.dev_patch_hash(self) } if dimg.dapp.dev_mode?
            end
          end

          def prepare_local_git_artifacts_command
            return super unless dimg.dapp.dev_mode?
            dimg.local_git_artifacts.each { |git_artifact| image.add_command git_artifact.apply_dev_patch_command(self) }
          end

          def layer_commit(git_artifact)
            commits[git_artifact] ||= begin
              git_artifact.latest_commit
            end
          end

          def empty?
            dependencies_empty? || git_artifacts_without_changes?
          end

          private

          def git_artifacts_without_changes?
            dimg.git_artifacts.all? do |git_artifact|
              from_commit = prev_g_a_stage.layer_commit(git_artifact)
              to_commit = dimg.dev_mode? ? nil : layer_commit(git_artifact)
              !git_artifact.any_changes?(from_commit, to_commit)
            end
          end

          def commit_list
            dimg.git_artifacts.map { |git_artifact| layer_commit(git_artifact) }
          end
        end # GALatestPatch
      end # Stage
    end # Build
  end # Dimg
end # Dapp
