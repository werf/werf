module Dapp
  module Dimg
    module Config
      module Directive
        module Docker
          class Base < Directive::Base
            attr_reader :_from, :_from_cache_version

            def from(image, cache_version: nil)
              sub_directive_eval do
                image = image.to_s
                raise(::Dapp::Error::Config, code: :docker_from_incorrect, data: { name: image }) unless ::Dapp::Dimg::Image::Stage.image_name?(image)
                @_from = image.include?(':') ? image : [image, 'latest'].join(':')
                @_from_cache_version = cache_version
              end
            end
          end
        end
      end
    end
  end
end
