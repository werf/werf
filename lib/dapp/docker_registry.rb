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

    def catalog
      resp_if_success(api_request('_catalog'))
    end

    def tags
      resp_if_success(api_request("#{repo}/tags/list"))['tags']
    end

    def image_id_by_tag(tag)
      resp = resp_if_success(api_request("#{repo}/manifests/#{tag}", headers: { Accept: 'application/vnd.docker.distribution.manifest.v2+json' }))
      resp['config']['digest']
    end

    def exist?
      catalog['repositories'].include?(repo)
    end

    protected

    def resp_if_success(resp)
      raise Error::Registry, code: :response_with_error_status unless resp.status == 200
      JSON.load(resp.body)
    end

    def api_request(uri, method: :get, **kwargs)
      connection.request(path: api_uri(uri), method: method, **kwargs)
    end

    def connection
      @connection ||= begin
        options = begin
          check_registry
          {}
        rescue Excon::Error::Unauthorized
          authorization_options
        rescue Excon::Error => e
          raise Error::Registry, e.net_status
        end
        Excon.new(api_url, **options)
      end
    end

    def check_registry
      check_url(registry_url, :not_found)
      check_url(api_url, :api_version_not_supported)
    end

    def check_url(url, code)
      raw_request(url, expects: [200])
    rescue Excon::Error::Unauthorized => _e
      raise
    rescue Excon::Error => _e
      raise Error::Registry, code: code
    end

    def raw_request(url, method: :get, **kwargs)
      Excon.new(url).request(method: method, **kwargs)
    end

    def authorization_options
      case authenticate_header = raw_request(api_url).headers['Www-Authenticate']
      when /Bearer/ then { headers: { Authorization: "Bearer #{authorization_token(authenticate_header)}" } }
      when /Basic/ then { headers: { Authorization: "Basic #{authorization_auth}" } }
      else raise Error::Registry, code: :authenticate_type_not_supported
      end
    end

    def authorization_token(authenticate_header)
      @token ||= begin
        options = parse_authenticate_header(authenticate_header)
        realm = options.delete(:realm)
        resp_if_success(raw_request(realm, headers: { Authorization: "Basic #{authorization_auth}" }, query: options))['token']
      end
    end

    def parse_authenticate_header(header)
      [:realm, :service, :scope].map do |option|
        /#{option}="(?<#{option}>[[^"].]*)/ =~ header
        next unless binding.local_variable_defined?(option)
        [option, binding.local_variable_get(option)]
      end.compact.to_h
    end

    def authorization_auth
      @authorization_auth ||= begin
        auths = auths_section_from_docker_config
        r = registry_with_repo
        loop do
          break unless r.include?('/') && !auths.keys.any? { |auth| auth.start_with?(r) }
          r = chomp_name(r)
        end
        credential = (auths[r] || auths.find { |repo, _| repo == r })
        raise Error::Registry, code: :user_not_authorized if credential.nil?
        credential['auth']
      end
    end

    def auths_section_from_docker_config
      @auths ||= begin
        file = Pathname(File.join(Dir.home, '.docker', 'config.json'))
        raise Error::Registry, code: :user_not_authorized unless file.exist?
        JSON.load(file.read)['auths'].tap { |auths| raise Error::Registry, code: :user_not_authorized if auths.nil? }
      end
    end

    def api_url
      File.join(registry_url, api_uri)
    end

    def api_uri(*uri)
      File.join(API_VERSION, '/', *uri)
    end

    private

    def chomp_name(r)
      r.split('/')[0..-2].join('/')
    end
  end
end # Dapp
