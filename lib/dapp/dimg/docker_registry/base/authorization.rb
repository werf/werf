module Dapp
  module Dimg
    module DockerRegistry
      class Base
        module Authorization
          def authorization_options(url)
            (@authorization_options ||= {})[@repo_suffix] ||= begin
              case authenticate_header = raw_request(url).headers['Www-Authenticate']
                when /Bearer/ then { headers: { Authorization: "Bearer #{authorization_token(authenticate_header)}" } }
                when /Basic/ then { headers: { Authorization: "Basic #{authorization_auth}" } }
                when nil then {}
                else raise Error::Registry, code: :authenticate_type_not_supported, data: { registry: api_url }
              end
            end
          end

          def authorization_token(authenticate_header)
            options = parse_authenticate_header(authenticate_header)
            realm = options.delete(:realm)
            begin
              response = raw_request(realm, headers: { Authorization: "Basic #{authorization_auth}" }, query: options, expects: [200])
            rescue Error::Registry
              raise unless (response = raw_request(realm, query: options)).status == 200
            end
            JSON.load(response.body)['token']
          end

          def parse_authenticate_header(header)
            [:realm, :service, :scope].map do |option|
              /#{option}="([[^"].]*)/ =~ header
              next unless Regexp.last_match(1)
              [option, Regexp.last_match(1)]
            end.compact.to_h
          end

          def authorization_auth
            auths = auths_section_from_docker_config
            r = repo
            loop do
              break unless r.include?('/') && !auths.keys.any? { |auth| auth.start_with?(r) }
              r = chomp_name(r)
            end
            credential = (auths[r] || auths.find { |repo, _| repo == r })
            user_not_authorized! if credential.nil?
            credential['auth']
          end

          def auths_section_from_docker_config
            file = Pathname(File.join(::Dapp::Dapp.host_docker_config_dir, 'config.json'))
            user_not_authorized! unless file.exist?
            JSON.load(file.read)['auths'].tap { |auths| user_not_authorized! if auths.nil? }
          end

          private

          def chomp_name(r)
            r.split('/')[0..-2].join('/')
          end
        end
      end # Base
    end # DockerRegistry
  end # Dimg
end # Dapp
