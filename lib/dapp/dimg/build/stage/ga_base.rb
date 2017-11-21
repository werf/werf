module Dapp
  module Dimg
    module Build
      module Stage
        class GABase < Base
          def prepare_image
            super do
              image.add_volumes_from dimg.dapp.gitartifact_container
              image.add_volume "#{dimg.tmp_path('archives')}:#{dimg.container_tmp_path('archives')}:ro"
              image.add_volume "#{dimg.tmp_path('patches')}:#{dimg.container_tmp_path('patches')}:ro"

              dimg.git_artifacts.each do |git_artifact|
                image.add_service_change_label("dapp-git-#{git_artifact.paramshash}-commit".to_sym => git_artifact.latest_commit)
                image.add_command git_artifact.send(apply_command_method, self)
              end
            end
          end

          def empty?
            dimg.git_artifacts.empty? || super
          end

          def g_a_stage?
            true
          end

          def image_should_be_untagged_condition
            return false unless image.tagged?
            dimg.git_artifacts.any? do |git_artifact|
              !git_artifact.repo.commit_exists? layer_commit(git_artifact)
            end
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
        end # GABase
      end # Stage
    end # Build
  end # Dimg
end # Dapp
