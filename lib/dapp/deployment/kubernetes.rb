module Dapp
  module Deployment
    class Kubernetes
      def initialize(namespace: nil)
        @namespace = namespace
        @query_parameters = {}
      end

      def namespace
        @namespace || 'default'
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
        '/apis/extensions/v1beta1' => [:deployment, :replicaset]
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

      def pod_log(name, follow: false, **query_parameters, &blk)
        excon_parameters = follow ? { response_block: blk } : {}
        request!(:get,
                 "/api/v1/namespaces/#{namespace}/pods/#{name}/log",
                 excon_parameters: excon_parameters,
                 **{ follow: follow }.merge(query_parameters))
      rescue Excon::Error::Timeout
        raise Error::Timeout
      end

      def event_list(**query_parameters)
        request!(:get, "/api/v1/namespaces/#{namespace}/events", **query_parameters)
      end

      protected

      # query_parameters — соответствует 'Query Parameters' в документации kubernetes
      # excon_parameters — соответствует опциям Excon
      # body — hash для http-body, соответствует 'Body Parameters' в документации kubernetes, опционален
      def request!(method, path, body: nil, excon_parameters: {}, **query_parameters)
        with_connection(excon_parameters: excon_parameters) do |conn|
          request_parameters = {method: method, path: path, query: @query_parameters.merge(query_parameters)}
          request_parameters[:body] = JSON.dump(body) if body
          load_body! conn.request(request_parameters), request_parameters
        end
      end

      def load_body!(response, request_parameters)
        if response.status.to_s.start_with? '5'
          raise Error::Base, code: :server_error, data: {
            response_http_status: response.status,
            response_http_body: response.body,
            request_parameters: request_parameters
          }
        elsif response.status.to_s == '404'
          raise Error::NotFound, data: {request_parameters: request_parameters}
        elsif not response.status.to_s.start_with? '2'
          raise Error::Base, code: :bad_request, data: {
            response_http_status: response.status,
            response_http_body: response.body,
            request_parameters: request_parameters
          }
        else
          JSON.load(response.body)
        end
      end

      def with_connection(excon_parameters: {}, &blk)
        old_ssl_ca_file = Excon.defaults[:ssl_ca_file]
        old_middlewares = Excon.defaults[:middlewares].dup

        begin
          Excon.defaults[:ssl_ca_file] = kube_config.fetch('clusters', [{}]).first.fetch('cluster', {}).fetch('certificate-authority', nil)
          Excon.defaults[:middlewares] << Excon::Middleware::RedirectFollower

          return yield Excon.new(kube_server_url, **kube_server_options(excon_parameters))
        ensure
          Excon.defaults[:ssl_ca_file] = old_ssl_ca_file
          Excon.defaults[:middlewares] = old_middlewares
        end
      end

      def kube_server_url
        @kube_server_url ||= begin
          kube_config.fetch('clusters', [{}]).first.fetch('cluster', {}).fetch('server', nil).tap do |url|
            begin
              Excon.new(url, **kube_server_options).get
            rescue Excon::Error::Socket
              raise Error::Base, code: :kube_server_connection_refused, data: { url: url }
            end
          end
        end
      end

      def kube_server_options(excon_parameters = {})
        { client_cert: kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-certificate', nil),
          client_key: kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-key', nil) }.merge(excon_parameters)
      end

      def kube_config
        @kube_config ||= begin
          if File.exist?((kube_config_path = File.join(ENV['HOME'], '.kube/config')))
            YAML.load_file(kube_config_path)
          else
            raise Error::Base, code: :kube_config_not_found, data: { path: kube_config_path }
          end
        end
      end
    end
  end
end
