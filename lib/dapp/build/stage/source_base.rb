module Dapp
  module Build
    module Stage
      # Base of source stages
      class SourceBase < Base
        attr_accessor :prev_source_stage, :next_source_stage

        GITARTIFACT_IMAGE = 'dappdeps/gitartifact:0.1.3'.freeze

        def prev_source_stage
          prev_stage.prev_stage
        end

        def next_source_stage
          next_stage.next_stage
        end

        def save_in_cache!
          super
          layers_commits_write!
        end

        def signature
          hashsum [dependencies_checksum, *commit_list]
        end

        def image
          super do |image|
            bash_commands = []
            volumes = []
            application.git_artifacts.each do |git_artifact|
              volumes << "#{git_artifact.repo.path}:#{git_artifact.repo.container_path}"
              bash_commands.concat(git_artifact.send(apply_command_method, self))
            end

            unless bash_commands.empty?
              image.add_volumes_from(gitartifact_container)
              image.add_volume(volumes)
              image.add_commands 'export PATH=/.dapp/deps/gitartifact/bin:$PATH', *bash_commands
            end
            yield image if block_given?
          end
        end

        def dependencies_checksum
          hashsum [prev_stage.signature, artifacts_signatures]
        end

        def layer_commit(git_artifact)
          commits[git_artifact] ||= begin
            if layer_commit_file_path(git_artifact).exist?
              layer_commit_file_path(git_artifact).read.strip
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

        def should_be_not_detailed?
          true
        end

        def apply_command_method
          :apply_patch_command
        end

        def commit_list
          application.git_artifacts.map { |git_artifact| layer_commit(git_artifact) }
        end

        def layers_commits_write!
          application.git_artifacts.each { |git_artifact| layer_commit_file_path(git_artifact).write(layer_commit(git_artifact)) }
        end

        def layer_commit_file_path(git_artifact)
          application.metadata_path git_artifact.filename ".#{name}.#{git_artifact.paramshash}.#{dependencies_checksum}.commit"
        end

        def dependency_files_checksum(regs)
          hashsum(regs.map { |reg| Dir[File.join(application.home_path, reg)].map { |f| File.read(f) if File.file?(f) } })
        end

        private

        def commits
          @commits ||= {}
        end
      end # SourceBase
    end # Stage
  end # Build
end # Dapp
