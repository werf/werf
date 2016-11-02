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
          super
          image.add_volumes_from dimg.project.gitartifact_container

          dimg.git_artifacts.each do |git_artifact|
            image.add_volume "#{git_artifact.repo.path}:#{git_artifact.repo.container_path}:ro"
            image.add_command git_artifact.send(apply_command_method, self)
          end
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
