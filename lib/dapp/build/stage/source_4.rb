module Dapp
  module Build
    module Stage
      class Source4 < Base
        def name
          :source_4
        end

        def prev_source_stage_name
          :source_3
        end

        def source_4_actual?
          build.git_artifact_list.map {|git_artifact| git_artifact.source_4_actual?}.all?
        end

        def source_4_commit_list
          build.git_artifact_list.map {|git_artifact| git_artifact.source_4_commit}
        end

        def signature
          if source_4_actual?
            build.stages[:app_setup].signature
          else
            hashsum [build.stages[:app_setup].signature, *source_4_commit_list]
          end
        end

        def git_artifact_signature
          hashsum build.stages[:app_setup].signature
        end

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              git_artifact.source_4_apply!(image)
            end
          end
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
