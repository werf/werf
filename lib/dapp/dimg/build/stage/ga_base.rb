module Dapp
  module Dimg
    module Build
      module Stage
        class GABase < Base
          def dependencies_stage
            prev_stage
          end

          def prepare_image
            super do
              image.add_volumes_from dimg.dapp.gitartifact_container
              image.add_volume "#{dimg.tmp_path('archives')}:#{dimg.container_tmp_path('archives')}:ro"
              image.add_volume "#{dimg.tmp_path('patches')}:#{dimg.container_tmp_path('patches')}:ro"

              dimg.git_artifacts.each { |git_artifact| image.add_command git_artifact.send(apply_command_method, self) }
            end
          end

          def empty?
            dependencies_stage.empty?
          end

          def g_a_stage?
            true
          end

          def layer_commit(git_artifact)
            commits[git_artifact] ||= begin
              if dependencies_stage && dependencies_stage.image.tagged?
                dependencies_stage.image.labels["dapp-git-#{git_artifact.paramshash}-commit"]
              else
                git_artifact.latest_commit
              end
            end
          end

          def renew
            @commits = {}
            super
          end

          protected

          def should_not_be_detailed?
            true
          end

          def ignore_log_commands?
            true
          end

          def apply_command_method
            :apply_patch_command
          end

          private

          def commits
            @commits ||= {}
          end
        end # GABase
      end # Stage
    end # Build
  end # Dimg
end # Dapp
