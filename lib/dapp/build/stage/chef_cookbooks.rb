module Dapp
  module Build
    module Stage
      # ChefCookbooks
      class ChefCookbooks < Base
        def initialize(application, next_stage)
          @prev_stage = Setup.new(application, self)
          super
        end

        def image
          super do |image|
            application.builder.chef_cookbooks(image)
          end
        end

        protected

        def dependencies
          application.builder.chef_cookbooks_checksum
        end
      end # ChefCookbooks
    end # Stage
  end # Build
end # Dapp
