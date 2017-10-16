module Dapp
  module Dimg
    module Build
      module Stage
        class GARelatedDependenciesBase < GADependenciesBase
          def dependencies
            @dependencies ||= dimg.stage_by_name(related_stage_name).context
          end

          def empty?
            dimg.stage_by_name(related_stage_name).empty? && super
          end

          def related_stage_name
            raise
          end
        end # GARelatedDependenciesBase
      end # Stage
    end # Build
  end # Dimg
end # Dapp
