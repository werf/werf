module Dapp
  # DockerRegistry
  module DockerRegistry
    # Base
    class Base
      include Mod::Request
      include Mod::Authorization

      API_VERSION = 'v2'.freeze

      attr_accessor :repo
      attr_accessor :hostname_url
      attr_accessor :repo_suffix

      def initialize(repo, hostname_url, repo_suffix)
        self.repo = repo
        self.hostname_url = hostname_url
        self.repo_suffix = repo_suffix
      end

      def repo_exist?
        tags.any?
      end

      def tags
        @tags ||= api_request(repo_suffix, 'tags/list')['tags']
      end

      def image_id_by_tag(tag)
        response = api_request(repo_suffix, "/manifests/#{tag}", headers: { Accept: 'application/vnd.docker.distribution.manifest.v2+json' })
        response['config']['digest']
      end

      protected

      def api_request(*uri, **options)
        JSON.load(request(api_url(*uri), expects: [200], **options).body)
      rescue Excon::Error::NotFound
        raise Error::Registry, code: :not_found
      end

      def api_url(*uri)
        File.join(hostname_url, API_VERSION, '/', *uri)
      end
    end
  end # DockerRegistry
end # Dapp
