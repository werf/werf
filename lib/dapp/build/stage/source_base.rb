module Dapp
  module Build
    module Stage
      # Base of source stages
      class SourceBase < Base
        attr_accessor :prev_source_stage, :next_source_stage

        GITARTIFACT_IMAGE = 'dappdeps/gitartifact:0.1.5'.freeze

        def prev_source_stage
          dependencies_stage.prev_stage.prev_stage
        end

        def next_source_stage
          next_stage.next_stage.next_stage
        end

        def dependencies_stage
          prev_stage
        end

        def image
          super do |image|
            image.add_volumes_from gitartifact_container
            image.add_command 'export PATH=/.dapp/deps/gitartifact/bin:$PATH'

            application.git_artifacts.each do |git_artifact|
              image.add_volume "#{git_artifact.repo.path}:#{git_artifact.repo.container_path}:ro"
              image.add_command git_artifact.send(apply_command_method, self)
            end
            yield image if block_given?
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

        def gitartifact_container_name # FIXME: hashsum(image) or dockersafe()
          GITARTIFACT_IMAGE.tr('/', '_').tr(':', '_')
        end

        def gitartifact_container
          @gitartifact_container ||= begin
            if application.shellout("docker inspect #{gitartifact_container_name}").exitstatus.nonzero?
              application.log_secondary_process(application.t(code: 'process.git_artifact_loading'), short: true) do
                application.shellout ['docker run',
                                      '--restart=no',
                                      "--name #{gitartifact_container_name}",
                                      "--volume /.dapp/deps/gitartifact #{GITARTIFACT_IMAGE}",
                                      '2>/dev/null'].join(' ')
              end
            end

            gitartifact_container_name
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
      end # SourceBase
    end # Stage
  end # Build
end # Dapp
