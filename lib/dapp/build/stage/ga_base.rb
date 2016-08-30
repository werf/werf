module Dapp
  module Build
    module Stage
      # GABase
      class GABase < Base
        attr_accessor :prev_g_a_stage, :next_g_a_stage

        GITARTIFACT_IMAGE = 'dappdeps/gitartifact:0.1.5'.freeze

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
          image.add_volumes_from g_a_container
          image.add_command 'export PATH=/.dapp/deps/gitartifact/bin:$PATH'

          application.git_artifacts.each do |git_artifact|
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

        def g_a_container_name # FIXME: hashsum(image) or dockersafe()
          GITARTIFACT_IMAGE.tr('/', '_').tr(':', '_')
        end

        def g_a_container
          @gitartifact_container ||= begin
            if application.project.shellout("docker inspect #{g_a_container_name}").exitstatus.nonzero?
              application.project.log_secondary_process(application.project.t(code: 'process.git_artifact_loading'), short: true) do
                application.project.shellout ['docker create',
                                              "--name #{g_a_container_name}",
                                              "--volume /.dapp/deps/gitartifact #{GITARTIFACT_IMAGE}"].join(' ')
              end
            end

            g_a_container_name
          end
        end

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
