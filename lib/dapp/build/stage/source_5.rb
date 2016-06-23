module Dapp
  module Build
    module Stage
      class Source5 < SourceBase
        def initialize(build)
          @prev_stage = Source4.new(build, self)
          @build = build
        end

        def name
          :source_5
        end

        def prev_source_stage
          prev_stage
        end

        def next_source_stage
          nil
        end

        def image
          super do |image|
            build.git_artifact_list.each do |git_artifact|
              git_artifact.layer_apply!(image, self)
            end
            image.build_opts!({ expose: build.conf[:exposes] }) unless build.conf[:exposes].nil?
          end
        end

        def signature
          source_4 = prev_stage
          if layers_actual?
            source_4.signature
          else
            hashsum [source_4.signature, *commit_list]
          end
        end

        def git_artifact_signature
          app_setup = prev_stage.prev_stage
          hashsum app_setup.signature
        end
      end # Source5
    end # Stage
  end # Build
end # Dapp
