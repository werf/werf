module Dapp
  module Build
    module Stage
      class Source4 < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = AppSetup.new(build, self)
          super
        end

        def name
          :source_4
        end

        def next_source_stage
          next_stage
        end

        def signature
          app_setup = prev_stage
          if layers_actual?
            app_setup.signature
          else
            hashsum [app_setup.signature, *commit_list]
          end
        end

        def git_artifact_signature
          hashsum prev_stage.signature
        end

        def layer_actual?(git_artifact)
          # FIXME git_artifact.patch_size(layer_commit(stage), layer_commit(stage.next_source_stage)) > NNN
          super and git_artifact.patch_size_valid?(next_source_stage)
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
