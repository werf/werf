module Dapp
  # DockerRegistry
  module DockerRegistry
    module Mod
      # Request
      module Request
        def request(url, **options)
          raw_request(url, deep_merge(options, authorization_options(url)))
        end

        def raw_request(url, **options)
          Excon.new(url).request(deep_merge(default_request_options, options))
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

        private

        def deep_merge(hash1, hash2)
          hash1.merge(hash2) do |_, v1, v2|
            if v1.is_a?(Hash) && v2.is_a?(Hash)
              v1.merge(v2)
            else
              [v1, v2].flatten
            end
          end
        end
      end
    end # Mod
  end # DockerRegistry
end # Dapp

Dapp::DockerRegistry::Mod::Request.extend Dapp::DockerRegistry::Mod::Request
