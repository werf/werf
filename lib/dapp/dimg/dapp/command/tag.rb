module Dapp
  module Dimg
    module Dapp
      module Command
        module Tag
          def tag(tag)
            one_dimg!
            validate_image_name!(tag)

            Dimg.new(config: build_configs.first, dapp: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
              app.tag!(tag)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
