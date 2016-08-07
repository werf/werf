module Dapp
  module Build
    module Stage
      # Source3
      class Source3 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = Source3Dependencies.new(application, self)
          super
        end

        def next_source_stage
          super.next_stage
        end
      end # Source3
    end # Stage
  end # Build
end # Dapp
