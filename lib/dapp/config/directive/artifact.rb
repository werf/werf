module Dapp
  module Config
    module Directive
      class Artifact < Directive::GitArtifactLocal
        attr_accessor :_before, :_after

        def before(stage)
          @_before = stage
        end

        def after(stage)
          @_after = stage
        end

        def _exports
          super do |export|
            export._before ||= @_before
            export._after ||= @_after
          end
        end

        protected

        class Export < Directive::GitArtifactLocal::Export
          attr_accessor :_before, :_after

          def before(stage)
            @_before = stage
          end

          def after(stage)
            @_after = stage
          end
        end
      end
    end
  end
end
