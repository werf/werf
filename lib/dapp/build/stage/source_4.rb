module Dapp
  module Build
    module Stage
      # Source4
      class Source4 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = Source4Dependencies.new(application, self)
          super
        end

        def prev_source_stage
          super.prev_stage
        end

        def next_source_stage
          next_stage
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
