module Dapp
  # DockerRegistry
  class DockerRegistry
    API_VERSION = 'v2'.freeze

    attr_reader :registry_with_repo
    attr_reader :registry_url
    attr_reader :repo

    def initialize(registry_with_repo)
      @registry_with_repo = registry_with_repo
      raise Error::Registry, code: :incorrect_repo if (parts = registry_with_repo.split('/')).one?
      @registry_url = File.join('http://', parts.shift)
      @repo = parts.join('/')
    end

    def repo_exist?
      catalog['repositories'].include?(repo)
    end

    def tags
      resp_if_success(api_request("#{repo}/tags/list"))['tags']
    end

    def repo_apps
      tags.select { |tag| !tag.start_with?('dappstage') }.map { |tag| [tag, image_id(tag)] }.to_h
    end

    protected

    def catalog
      resp_if_success api_request('_catalog')
    end

    def image_id(tag)
      resp = resp_if_success(api_request("#{repo}/manifests/#{tag}", headers: { Accept: 'application/vnd.docker.distribution.manifest.v2+json' }))
      resp['config']['digest']
    end

    def raw_connection
      Excon.new("#{File.join(registry_url, API_VERSION)}/")
    end

    def api_request(uri, method: :get, **kwargs)
      connection.request(path: File.join(API_VERSION, uri), method: method, **kwargs)
    end

    def connection
      @connection ||= begin
        check_registry
      rescue Excon::Error::OK
        Excon.new(registry_url)
      rescue Excon::Error::Unauthorized
        options = authorization_options || {}
        Excon.new(registry_url, **options)
      rescue StandardError => e
        raise Error::Registry, e.net_status
      end
    end

    def check_registry
      raise Error::Registry, :not_found if Excon.new(registry_url).request(expects: [404], method: :get)
    rescue Excon::Error::OK
      raise Error::Registry, :api_version_not_supported if raw_connection.request(expects: [404], method: :get)
    end

    def authorization_options
      case raw_connection.request(expects: [401], method: :get).headers['Www-Authenticate']
      when /Bearer/ then raise # TODO
      when /Basic/ then { headers: { Authorization: "Basic #{auth}" } }
      else raise Error::Registry, :authenticate_type_not_supported
      end
    end

    def auth
      @auth ||= begin
        r = registry_with_repo
        loop do
          break unless r.include?('/') && !auths.keys.any? { |auth| auth.start_with?(r) }
          r = chomp_name(r)
        end
        raise Error::Registry, :user_not_authorized if (credential = (auths[r] || auths.find { |repo, _| repo == r })).nil?
        credential['auth']
      end
    end

    def auths
      @auths ||= begin
        file = Pathname(File.join(Dir.home, '.docker', 'config.json'))
        raise Error::Registry, :user_not_authorized unless file.exist?
        JSON.load(file.read)['auths'].tap { |auths| raise Error::Registry, :user_not_authorized if auths.nil? }
      end
    end

    def resp_if_success(resp)
      raise Error::Registry, :response_with_error_status unless resp.status == 200
      JSON.load(resp.body)
    end

    private

    def chomp_name(r)
      r.split('/')[0..-2].join('/')
    end
  end
end # Dapp
