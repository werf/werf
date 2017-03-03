module Dapp
  module Dimg
    class Dapp
      module Command
        module Tag
          def tag(tag)
            one_dimg!
            raise Error::Command, code: :tag_command_incorrect_tag, data: { name: tag } unless ::Dapp::Dimg::Image::Docker.image_name?(tag)
            Dimg.new(config: build_configs.first, dapp: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
              app.tag!(tag)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
