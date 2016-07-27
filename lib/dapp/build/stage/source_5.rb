module Dapp
  module Build
    module Stage
      # Source5
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

        def signature
          hashsum [super, change_options]
        end

        def image
          super do |image|
            change_options.each do |k, v|
              next if v.nil? || v.empty?
              image.public_send("add_change_#{k}", v)
            end
          end
        end

        protected

        def change_options
          application.config._docker._change_options
        end

        def layers_commits_write!
          nil
        end
      end # Source5
    end # Stage
  end # Build
end # Dapp
