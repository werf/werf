module Dapp
  module Build
    module Stage
      # Source2
      class Source2 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = Source2Dependencies.new(application, self)
          super
        end
      end # Source2
    end # Stage
  end # Build
end # Dapp
