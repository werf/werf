module Dapp
  module Build
    module Stage
      class SourceBase < Base
        attr_accessor :prev_source_stage, :next_source_stage
        MAX_PATCH_SIZE = 1024*1024

        def prev_source_stage
          prev_stage.prev_stage
        end

        def next_source_stage
          next_stage.next_stage
        end

        def signature
          hashsum [dependencies_checksum, *commit_list]
        end

        def dependencies_checksum
          hashsum [prev_stage.signature]
        end

        def layer_commit(git_artifact)
          if layer_commit_file_path(git_artifact).exist?
            layer_commit_file_path(git_artifact).read.strip
          else
            layer_commit_write!(git_artifact)
            git_artifact.repo_latest_commit
          end
        end

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              layer_commit_write!(git_artifact)
              layer_timestamp_write!(git_artifact)
              image.add_commands git_artifact.send(apply_command_method, self)
            end
            yield image if block_given?
          end
        end

        protected

        def apply_command_method
          :apply_patch_command
        end

        def commit_list
          build.git_artifact_list.map { |git_artifact| layer_commit(git_artifact) }
        end

        def patch_size_valid?(git_artifact)
          git_artifact.patch_size(prev_source_stage.layer_commit(git_artifact), layer_commit(git_artifact)) < MAX_PATCH_SIZE
        end

        def layer_timestamp(git_artifact)
          if layer_timestamp_file_path(git_artifact).exist?
            layer_timestamp_file_path(git_artifact).read.strip.to_i
          else
            layer_timestamp_write!(git_artifact)
            git_artifact.repo.commit_at(layer_commit(git_artifact)).to_i
          end
        end

        def layer_commit_write!(git_artifact)
          git_artifact.file_atomizer.add_path(layer_commit_file_path(git_artifact))
          layer_commit_file_path(git_artifact).write(git_artifact.repo_latest_commit + "\n")
        end

        def layer_timestamp_write!(git_artifact)
          git_artifact.file_atomizer.add_path(layer_timestamp_file_path(git_artifact))
          layer_timestamp_file_path(git_artifact).write("#{git_artifact.repo.commit_at(layer_commit(git_artifact)).to_i}\n")
        end

        def layer_commit_filename(git_artifact)
          layer_filename git_artifact, '.commit'
        end

        def layer_timestamp_filename(git_artifact)
          layer_filename git_artifact, '.timestamp'
        end

        def layer_commit_file_path(git_artifact)
          build_path layer_commit_filename(git_artifact)
        end

        def layer_timestamp_file_path(git_artifact)
          build_path layer_timestamp_filename(git_artifact)
        end

        def layer_filename(git_artifact, ending)
          git_artifact.filename ".#{name}.#{git_artifact.paramshash}.#{dependencies_checksum}#{ending}"
        end

        def build_path(*path)
          build.build_path(*path)
        end

        def container_build_path(*path)
          build.container_build_path(*path)
        end
      end # SourceBase
    end # Stage
  end # Build
end # Dapp
