module Dapp
  module Kube
    module Kubernetes
      class Client
        include Helper::YAML

        def initialize(namespace: nil)
          @namespace = namespace
          @query_parameters = {}
        end

        def namespace
          @namespace || kube_context_config['context']['namespace'] || 'default'
        end

        # Чтобы не перегружать методы явной передачей namespace.
        # Данный метод может пригодиться только в ситуации, когда надо указать другой namespace,
        # в большинстве случаев используется namespace из конструктора.
        def with_namespace(namespace, &blk)
          old_namespace = @namespace
          begin
            @namespace = namespace
            return yield
          ensure
            @namespace = old_namespace
          end
        end

        def with_query(query, &blk)
          old_query = @query_parameters
          begin
            @query_parameters = query
            return yield
          ensure
            @query_parameters = old_query
          end
        end

        # NOTICE: Название метода аналогично kind'у выдаваемого результата.
        # NOTICE: В данном случае в результате kind=DeploymentList.
        # NOTICE: Методы создания/обновления/удаления сущностей kubernetes заканчиваются на '!'. Например, create_deployment!.

        {
          '/api/v1' => [:service, :replicationcontroller, :pod],
          '/apis/extensions/v1beta1' => [:deployment, :replicaset],
          '/apis/batch/v1' => [:job]
        }.each do |api, objects|
          objects.each do |object|
            define_method :"#{object}_list" do |**query_parameters|
              request!(:get, "#{api}/namespaces/#{namespace}/#{object}s", **query_parameters)
            end

            define_method object do |name, **query_parameters|
              request!(:get, "#{api}/namespaces/#{namespace}/#{object}s/#{name}", **query_parameters)
            end

            define_method "#{object}_status" do |name, **query_parameters|
              request!(:get, "#{api}/namespaces/#{namespace}/#{object}s/#{name}/status", **query_parameters)
            end

            define_method :"create_#{object}!" do |spec, **query_parameters|
              request!(:post, "#{api}/namespaces/#{namespace}/#{object}s", body: spec, **query_parameters)
            end

            define_method :"replace_#{object}!" do |name, spec, **query_parameters|
              request!(:put, "#{api}/namespaces/#{namespace}/#{object}s/#{name}", body: spec, **query_parameters)
            end

            define_method :"delete_#{object}!" do |name, **query_parameters|
              request!(:delete, "#{api}/namespaces/#{namespace}/#{object}s/#{name}", **query_parameters)
            end

            define_method :"delete_#{object}s!" do |**query_parameters|
              request!(:delete, "#{api}/namespaces/#{namespace}/#{object}s", **query_parameters)
            end

            define_method :"#{object}?" do |name, **query_parameters|
              public_send(:"#{object}_list", **query_parameters)['items'].map { |item| item['metadata']['name'] }.include?(name)
            end
          end
        end

        def namespace_list(**query_parameters)
          request!(:get, '/api/v1/namespaces', **query_parameters)
        end

        def namespace?(name, **query_parameters)
          namespace_list(**query_parameters)['items'].map { |item| item['metadata']['name'] }.include?(name)
        end

        def create_namespace!(name, **query_parameters)
          request!(:post, '/api/v1/namespaces', body: { metadata: { name: name } }, **query_parameters)
        end

        def delete_namespace!(name, **query_parameters)
          request!(:delete, "/api/v1/namespaces/#{name}", **query_parameters)
        end

        def pod_log(name, follow: false, **query_parameters, &blk)
          excon_parameters = follow ? { response_block: blk } : {}
          request!(:get,
                   "/api/v1/namespaces/#{namespace}/pods/#{name}/log",
                   excon_parameters: excon_parameters,
                   response_body_parameters: {json: false},
                   **{ follow: follow }.merge(query_parameters))
        rescue Excon::Error::Timeout
          raise Error::Timeout
        rescue Error::Base => err
          if err.net_status[:code] == :bad_request and err.net_status[:data][:response_body]
            msg = err.net_status[:data][:response_body]['message']
            if msg.end_with? 'ContainerCreating'
              raise Error::Pod::ContainerCreating, data: err.net_status[:data]
            elsif msg.end_with? 'PodInitializing'
              raise Error::Pod::PodInitializing, data: err.net_status[:data]
            end
          end

          raise
        end

        def event_list(**query_parameters)
          request!(:get, "/api/v1/namespaces/#{namespace}/events", **query_parameters)
        end

        protected

        # query_parameters — соответствует 'Query Parameters' в документации kubernetes
        # excon_parameters — соответствует опциям Excon
        # body — hash для http-body, соответствует 'Body Parameters' в документации kubernetes, опционален
        def request!(method, path, body: nil, excon_parameters: {}, response_body_parameters: {}, **query_parameters)
          with_connection(excon_parameters: excon_parameters) do |conn|
            request_parameters = {method: method, path: path, query: @query_parameters.merge(query_parameters)}
            request_parameters[:body] = JSON.dump(body) if body
            load_body! conn.request(request_parameters), request_parameters, **response_body_parameters
          end
        end

        def load_body!(response, request_parameters, json: true)
          response_ok = response.status.to_s.start_with? '2'

          if response_ok
            if json
              JSON.parse(response.body)
            else
              response.body
            end
          else
            err_data = {}
            err_data[:response_http_status] = response.status
            if response_body = (JSON.parse(response.body) rescue nil)
              err_data[:response_body] = response_body
            else
              err_data[:response_raw_body] = response.body
            end
            err_data[:request_parameters] = request_parameters

            if response.status.to_s.start_with? '5'
              raise Error::Base, code: :server_error, data: err_data
            elsif response.status.to_s == '404'
              case err_data.fetch(:response_body, {}).fetch('details', {})['kind']
              when 'pods'
                raise Error::Pod::NotFound, data: err_data
              else
                raise Error::NotFound, data: err_data
              end
            elsif not response.status.to_s.start_with? '2'
              raise Error::Base, code: :bad_request, data: err_data
            end
          end
        end

        def with_connection(excon_parameters: {}, &blk)
          old_ssl_ca_file = Excon.defaults[:ssl_ca_file]
          old_ssl_cert_store = Excon.defaults[:ssl_cert_store]
          old_middlewares = Excon.defaults[:middlewares].dup

          begin
            ssl_cert_store = OpenSSL::X509::Store.new
            if ssl_ca_file = kube_config.fetch('clusters', [{}]).first.fetch('cluster', {}).fetch('certificate-authority', nil)
              ssl_cert_store.add_file ssl_ca_file
            elsif ssl_ca_data = kube_config.fetch('clusters', [{}]).first.fetch('cluster', {}).fetch('certificate-authority-data', nil)
              ssl_cert_store.add_cert OpenSSL::X509::Certificate.new(Base64.decode64(ssl_ca_data))
            end

            Excon.defaults[:ssl_ca_file] = nil
            Excon.defaults[:ssl_cert_store] = ssl_cert_store
            Excon.defaults[:middlewares] << Excon::Middleware::RedirectFollower

            connection = begin
              Excon.new(kube_cluster_config['cluster']['server'], **kube_server_options(excon_parameters)).tap(&:get)
            rescue Excon::Error::Socket => err
              raise Error::ConnectionRefused,
                code: :kube_server_connection_refused,
                data: { kube_cluster_config: kube_cluster_config, kube_user_config: kube_user_config, error: err.message }
            end

            return yield connection
          ensure
            Excon.defaults[:ssl_ca_file] = old_ssl_ca_file
            Excon.defaults[:ssl_cert_store] = old_ssl_cert_store
            Excon.defaults[:middlewares] = old_middlewares
          end
        end

        def kube_server_options(excon_parameters = {})
          {}.tap do |opts|
            client_cert = kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-certificate', nil)
            opts[:client_cert] = client_cert if client_cert

            client_cert_data = kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-certificate-data', nil)
            opts[:client_cert_data] = Base64.decode64(client_cert_data) if client_cert_data

            client_key = kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-key', nil)
            opts[:client_key] = client_key if client_key

            client_key_data = kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-key-data', nil)
            opts[:client_key_data] = Base64.decode64(client_key_data) if client_key_data
          end
        end

        def kube_user_config
          @kube_user_config ||= begin
            kube_config.fetch('users', []).find {|user| user['name'] == kube_context_config['context']['user']} || begin
              raise Error::BadConfig, code: :kube_user_not_found, data: {context: kube_context_config}
            end
          end
        end

        def kube_cluster_config
          @kube_cluster_config ||= begin
            kube_config.fetch('clusters', []).find {|cluster| cluster['name'] == kube_context_config['context']['cluster']} || begin
              raise Error::BadConfig, code: :kube_cluster_not_found, data: {context: kube_context_config}
            end
          end
        end

        def kube_context_config
          @kube_context_config ||= begin
            kube_config.fetch('contexts', []).find {|context| context['name'] == kube_context_name} || begin
              raise Error::BadConfig, code: :kube_context_not_found, data: {context_name: kube_context_name}
            end
          end
        end

        def kube_context_name
          @kube_context_name ||= kube_config['current-context'] || begin
            if context = kube_config.fetch('contexts', []).first
              warn "[WARN] .kube/config current-context is not set, using context '#{context['name']}'"
              context['name']
            end
          end
        end

        def kube_config
          @kube_config ||= begin
            if File.exist?((kube_config_path = File.join(ENV['HOME'], '.kube/config')))
              yaml_load_file(kube_config_path)
            else
              raise Error::Base, code: :kube_config_not_found, data: { path: kube_config_path }
            end
          end
        end
      end # Client
    end # Kubernetes
  end # Kube
end # Dapp
