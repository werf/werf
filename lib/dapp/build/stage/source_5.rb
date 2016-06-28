module Dapp
  module Build
    module Stage
      class Source5 < SourceBase
        def initialize(application)
          @prev_stage = Source4.new(application, self)
          @application = application
        end

        def prev_source_stage
          prev_stage
        end

        def next_source_stage
          nil
        end

        def image
          super do |image|
            exposes = application.conf[:exposes]
            image.add_expose(exposes) unless exposes.nil?
          end
        end

        protected

        def layers_commits_write!
          nil
        end
      end # Source5
    end # Stage
  end # Build
end # Dapp
