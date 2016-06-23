module Dapp
  module Build
    module Stage
      class Source1Archive < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = InfraInstall.new(build, self)
          super
        end

        def name
          :source_1_archive
        end

        def prev_source_stage
          nil
        end

        def next_source_stage
          next_stage
        end

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              git_artifact.archive_apply!(image, self)
            end
          end
        end

        def signature
          hashsum [prev_stage.signature, *commit_list]
        end

        def git_artifact_signature
          hashsum prev_stage.signature
        end
      end # Source1Archive
    end # Stage
  end # Build
end # Dapp
