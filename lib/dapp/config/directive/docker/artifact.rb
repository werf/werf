module Dapp
  module Config
    module Directive
      # Docker
      module Docker
        class Artifact
          attr_reader :_from
          attr_reader :_from_cache_version

          def from(image, cache_version: nil)
            raise(Error::Config, code: :docker_from_incorrect, data: { name: image }) unless image =~ /^[[^ ].]+:[[^ ].]+$/
            @_from = image
            @_from_cache_version = cache_version
          end

          protected

          def clone
            Marshal.load(Marshal.dump(self))
          end
        end
      end
    end
  end
end
