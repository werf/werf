module Dapp
  module Dimg
    module DockerRegistry
      module Error
        class Base < ::Dapp::Dimg::Error::Registry
          def initialize(**net_status)
            super(**net_status, context: :registry)
          end
        end

        class ManifestInvalid < Base
          def initialize(url, registry, response_body)
            super(code: :manifest_invalid, data: {url: url, registry: registry, response_body: response_body})
          end
        end

        class ImageNotFound < Base
          def initialize(url, registry)
            super(code: :page_not_found, data: { url: url, registry: registry })
          end
        end
      end # Error
    end # DockerRegistry
  end # Dimg
end # Dapp
