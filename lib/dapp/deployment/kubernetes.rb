module Dapp
  module Deployment
    class Kubernetes
      class Error::NotFound < Error::Kubernetes; end

      def initialize(namespace: nil)
        @namespace = namespace
        @query_parameters = {}
        @kube_config = YAML.load_file(File.join(ENV['HOME'], '.kube/config'))
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

      # v1
      [:service, :replicationcontroller, :pod].each do |object|
        define_method :"#{object}_list" do |**query_parameters|
          request!(:get, "/api/v1/namespaces/#{namespace}/#{object}s", **query_parameters)
        end

        define_method object do |name, **query_parameters|
          request!(:get, "/api/v1/namespaces/#{namespace}/#{object}s/#{name}", **query_parameters)
        end

        define_method :"create_#{object}!" do |spec, **query_parameters|
          request!(:post, "/api/v1/namespaces/#{namespace}/#{object}s", body: spec, **query_parameters)
        end

        define_method :"replace_#{object}!" do |name, spec, **query_parameters|
          request!(:put, "/api/v1/namespaces/#{namespace}/#{object}s/#{name}", body: spec, **query_parameters)
        end

        define_method :"delete_#{object}!" do |name, **query_parameters|
          request!(:delete, "/api/v1/namespaces/#{namespace}/#{object}s/#{name}", **query_parameters)
        end

        define_method :"#{object}?" do |name, **query_parameters|
          public_send(:"#{object}_list", **query_parameters)['items'].map { |item| item['metadata']['name'] }.include?(name)
        end
      end

      # v1beta1
      [:deployment].each do |object|
        define_method :"#{object}_list" do |**query_parameters|
          request!(:get, "/apis/extensions/v1beta1/namespaces/#{namespace}/#{object}s", **query_parameters)
        end

        define_method object do |name, **query_parameters|
          request!(:get, "/apis/extensions/v1beta1/namespaces/#{namespace}/#{object}s/#{name}", **query_parameters)
        end

        define_method :"create_#{object}!" do |spec, **query_parameters|
          request!(:post, "/apis/extensions/v1beta1/namespaces/#{namespace}/#{object}s", body: spec, **query_parameters)
        end

        define_method :"replace_#{object}!" do |name, spec, **query_parameters|
          request!(:put, "/apis/extensions/v1beta1/namespaces/#{namespace}/#{object}s/#{name}", body: spec, **query_parameters)
        end

        define_method :"delete_#{object}!" do |name, **query_parameters|
          request!(:delete, "/apis/extensions/v1beta1/namespaces/#{namespace}/#{object}s/#{name}", **query_parameters)
        end

        define_method :"#{object}?" do |name, **query_parameters|
          public_send(:"#{object}_list", **query_parameters)['items'].map { |item| item['metadata']['name'] }.include?(name)
        end
      end

      protected

      # query_parameters — соответствует 'Query Parameters' в документации kubernetes
      # body — hash для http-body, соответствует 'Body Parameters' в документации kubernetes, опционален
      def request!(method, path, body: nil, **query_parameters)
        with_connection do |conn|
          request_parameters = {method: method, path: path, query: @query_parameters.merge(query_parameters)}
          request_parameters[:body] = JSON.dump(body) if body
          load_body! conn.request(request_parameters)
        end
      end

      def load_body!(response)
        if response.status.to_s.start_with? '5'
          raise Error::Kubernetes, code: :server_error, data: {http_status: response.status, http_body: response.body}
        else
          body = JSON.load(response.body)
          raise Error::NotFound if response.status.to_s == '404'
          raise Error::Kubernetes, code: :bad_request, data: {body: body} unless response.status.to_s.start_with? '2'
          body
        end
      end

      def with_connection(&blk)
        old_ssl_ca_file = Excon.defaults[:ssl_ca_file]
        old_middlewares = Excon.defaults[:middlewares].dup

        begin
          Excon.defaults[:ssl_ca_file] = @kube_config.fetch('clusters', [{}]).first.fetch('cluster', {}).fetch('certificate-authority', nil)
          Excon.defaults[:middlewares] << Excon::Middleware::RedirectFollower

          return yield Excon.new(
            @kube_config.fetch('clusters', [{}]).first.fetch('cluster', {}).fetch('server', nil),
            client_cert: @kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-certificate', nil),
            client_key: @kube_config.fetch('users', [{}]).first.fetch('user', {}).fetch('client-key', nil)
          )
        ensure
          Excon.defaults[:ssl_ca_file] = old_ssl_ca_file
          Excon.defaults[:middlewares] = old_middlewares
        end
      end
    end
  end
end
