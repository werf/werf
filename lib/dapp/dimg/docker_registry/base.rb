module Dapp
  module Dimg
    module DockerRegistry
      class Base
        include Request
        include Authorization

        API_VERSION = 'v2'.freeze

        attr_accessor :repo
        attr_accessor :hostname_url
        attr_accessor :repo_suffix

        def initialize(repo, hostname_url, repo_suffix)
          self.repo = repo
          self.hostname_url = hostname_url
          self.repo_suffix = repo_suffix
        end

        def tags
          api_request(repo_suffix, 'tags/list')['tags'] || []
        end

        def image_id(tag)
          response = manifest_v2(tag)
          response['config']['digest'] if response['schemaVersion'] == 2
        end

        def image_parent_id(tag)
          image_history(tag)['container_config']['Image']
        end

        def image_labels(tag)
          image_history(tag)['config']['Labels']
        end

        def image_delete(tag)
          image_blobs(tag).each { |hash| blob_delete(hash.values.first) }
          api_request(repo_suffix, "/manifests/#{image_digest(tag)}",
                      method: :delete,
                      expects: [202, 404],
                      headers: { Accept: 'Accept: application/vnd.docker.distribution.manifest.v2+json' })
        end

        def image_history(tag)
          response = manifest_v1(tag)
          JSON.load(response['history'].first['v1Compatibility'])
        end

        protected

        def image_digest(tag)
          raw_api_request(repo_suffix, "/manifests/#{tag}",
                          headers: { Accept: 'application/vnd.docker.distribution.manifest.v2+json' }).headers['Docker-Content-Digest']
        end

        def image_blobs(tag)
          manifest_v1(tag)['fsLayers']
        end

        def manifest_v1(tag)
          api_request(repo_suffix, "/manifests/#{tag}")
        end

        def manifest_v2(tag)
          api_request(repo_suffix, "/manifests/#{tag}",
                      headers: { Accept: 'application/vnd.docker.distribution.manifest.v2+json' })
        end

        def blob_delete(id)
          api_request(repo_suffix, "/blobs/#{id}",
                      method: :delete, expects: [202, 404])
        end

        def api_request(*uri, **options)
          JSON.load(raw_api_request(*uri, **options).body)
        end

        def raw_api_request(*uri, **options)
          url = api_url(*uri)
          request(url, **default_api_options.merge(options))
        rescue Excon::Error::MethodNotAllowed
          raise Error::Registry, code: :method_not_allowed, data: { url: url, registry: api_url, method: options[:method] }
        rescue Excon::Error::NotFound
          raise Error::Registry, code: :page_not_found, data: { url: url, registry: api_url }
        rescue Excon::Error::Unauthorized
          user_not_authorized!
        end

        def api_url(*uri)
          File.join(hostname_url, API_VERSION, '/', *uri)
        end

        def default_api_options
          { expects: [200] }
        end

        def user_not_authorized!
          raise Error::Registry, code: :user_not_authorized, data: { registry: api_url }
        end
      end
    end # DockerRegistry
  end # Dimg
end # Dapp
