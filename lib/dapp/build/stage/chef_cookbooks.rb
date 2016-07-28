module Dapp
  module Build
    module Stage
      # ChefCookbooks
      class ChefCookbooks < Base
        def initialize(application, next_stage)
          @prev_stage = AppSetup.new(application, self)
          super
        end

        def signature
          hashsum [prev_stage.signature, *application.builder.chef_cookbooks_checksum]
        end

        def image
          super do |image|
            application.builder.chef_cookbooks(image)
          end
        end
      end # ChefCookbooks
    end # Stage
  end # Build
end # Dapp
