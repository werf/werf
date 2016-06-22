module Dapp
  module Build
    module Stage
      class Source5 < Base
        def name
          :source_5
        end

        def prev_source_stage_name
          :source_4
        end

        def source_5_actual?
          build.git_artifact_list.map {|git_artifact| git_artifact.source_5_actual?}.all?
        end

        def source_5_commit
          build.git_artifact_list.map {|git_artifact| git_artifact.source_5_commit}.reduce(:+)
        end

        def signature
          if source_5_actual?
            build.stages[:source_4].signature
          else
            hashsum [build.stages[:source_4].signature, source_5_commit]
          end
        end

        def git_artifact_signature
          hashsum build.stages[:app_setup].signature
        end

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              git_artifact.source_5_apply!(image)
            end
          end
        end
      end # Source5
    end # Stage
  end # Build
end # Dapp
