module Dapp
  module Dimg
    module DockerRegistry
      class Base
        module Request
          def self.included(base)
            base.extend(self)
          end

          def request(url, **options)
            raw_request(url, options.in_depth_merge(authorization_options(url, method: options[:method] || :get)))
          end

          def raw_request(url, **options)
            Excon.new(url).request(default_request_options.in_depth_merge(options))
          end

          def url_available?(url)
            raw_request(url, expects: [200])
            true
          rescue Excon::Error
            false
          end

          protected

          def default_request_options
            { method: :get, omit_default_port: true }
          end
        end
      end # Base
    end # DockerRegistry
  end # Dimg
end # Dapp
