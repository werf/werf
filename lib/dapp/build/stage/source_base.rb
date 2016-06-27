module Dapp
  module Build
    module Stage
      class SourceBase < Base
        attr_accessor :prev_source_stage, :next_source_stage

        def prev_source_stage
          prev_stage.prev_stage
        end

        def next_source_stage
          next_stage.next_stage
        end

        def fixate!
          super
          layer_commits_write!
        end

        def signature
          hashsum [dependencies_checksum, *commit_list]
        end

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              layer_commit_change(git_artifact)
              image.add_commands git_artifact.send(apply_command_method, self)
            end
            yield image if block_given?
          end
        end

        def dependencies_checksum
          hashsum [prev_stage.signature]
        end

        def layer_commit(git_artifact)
          commits[git_artifact] ||= begin
            if layer_commit_file_path(git_artifact).exist?
              layer_commit_file_path(git_artifact).read.strip
            else
              git_artifact.repo_latest_commit
            end
          end
        end

        protected

        def apply_command_method
          :apply_patch_command
        end

        def commit_list
          build.git_artifact_list.map { |git_artifact| layer_commit(git_artifact) }
        end

        def layer_commits_write!
          build.git_artifact_list.each { |git_artifact| layer_commit_file_path(git_artifact).write(layer_commit(git_artifact)) }
        end

        def layer_commit_change(git_artifact)
          commits[git_artifact] ||= git_artifact.repo_latest_commit
        end

        def layer_commit_file_path(git_artifact)
          build_path git_artifact.filename ".#{name}.#{git_artifact.paramshash}.#{dependencies_checksum}.commit"
        end

        def build_path(*path)
          build.build_path(*path)
        end

        def container_build_path(*path)
          build.container_build_path(*path)
        end

        private

        def commits
          @commits ||= {}
        end
      end # SourceBase
    end # Stage
  end # Build
end # Dapp
