module Dapp
  module Build
    module Stage
      # Source1
      class Source1 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = Source1Dependencies.new(application, self)
          super
        end

        def prev_source_stage
          dependencies_stage.prev_stage
        end
      end # Source1
    end # Stage
  end # Build
end # Dapp
