module Dapp
  module Dimg
    module Config
      module Directive
        module Docker
          # Base
          class Base < Directive::Base
            attr_reader :_from, :_from_cache_version

            def from(image, cache_version: nil)
              image = image.to_s
              raise(Error::Config, code: :docker_from_incorrect,
                                   data: { name: image }) unless ::Dapp::Dimg::Image::Docker.image_name?(image)
              raise(Error::Config, code: :docker_from_without_tag, data: { name: image }) unless image.include?(':')
              @_from = image
              @_from_cache_version = cache_version
            end
          end
        end
      end
    end
  end
end
