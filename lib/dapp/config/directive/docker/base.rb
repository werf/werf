module Dapp
  module Config
    module Directive
      module Docker
        # Base
        class Base < Directive::Base
          attr_reader :_from, :_from_cache_version

          def from(image, cache_version: nil)
            raise(Error::Config, code: :docker_from_incorrect, data: { name: image }) unless image =~ /^[[^ ].]+:[[^ ].]+$/
            @_from = image
            @_from_cache_version = cache_version
          end
        end
      end
    end
  end
end
