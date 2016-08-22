module Dapp
  module Build
    module Stage
      module SetupGroup
        # ChefCookbooks
        class ChefCookbooks < Base
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = Setup.new(application, self)
            super
          end

          def dependencies
            [application.builder.chef_cookbooks_checksum]
          end

          def image
            super do |image|
              application.builder.chef_cookbooks(image)
            end
          end

          protected

          def should_be_not_detailed?
            true
          end

          def ignore_log_commands?
            true
          end
        end # ChefCookbooks
      end
    end # Stage
  end # Build
end # Dapp
