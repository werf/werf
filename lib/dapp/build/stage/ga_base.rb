module Dapp
  module Build
    module Stage
      # GABase
      class GABase < Base
        attr_accessor :prev_g_a_stage, :next_g_a_stage

        def prev_g_a_stage
          dependencies_stage.prev_stage.prev_stage
        end

        def next_g_a_stage
          next_stage.next_stage.next_stage
        end

        def dependencies_stage
          prev_stage
        end

        def prepare_image
          super do
            image.add_volumes_from dimg.dapp.gitartifact_container
            image.add_volume "#{dimg.tmp_path('archives')}:#{dimg.container_tmp_path('archives')}:ro"
            image.add_volume "#{dimg.tmp_path('patches')}:#{dimg.container_tmp_path('patches')}:ro"

            prepare_local_git_artifacts_command
            prepare_remote_git_artifacts_command
          end
        end
        
        def prepare_local_git_artifacts_command
          prepare_base_git_artifacts_command(dimg.local_git_artifacts)
        end

        def prepare_remote_git_artifacts_command
          prepare_base_git_artifacts_command(dimg.remote_git_artifacts)
        end

        def prepare_base_git_artifacts_command(git_artifacts)
          git_artifacts.each { |git_artifact| image.add_command git_artifact.send(apply_command_method, self) }
        end

        def empty?
          dependencies_stage.empty?
        end

        def layer_commit(git_artifact)
          commits[git_artifact] ||= begin
            if dependencies_stage && dependencies_stage.image.tagged?
              dependencies_stage.image.labels[git_artifact.full_name]
            else
              git_artifact.latest_commit
            end
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

        private

        def commits
          @commits ||= {}
        end
      end # GABase
    end # Stage
  end # Build
end # Dapp
