module Dapp
  module Deployment
    class Kubernetes
      class Error::NotFound < Error::Kubernetes; end

      def initialize(namespace: nil)
        @namespace = namespace
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

      # NOTICE: Название метода аналогично kind'у выдаваемого результата.
      # NOTICE: В данном случае в результате kind=DeploymentList.
      # NOTICE: Методы создания/обновления/удаления сущностей kubernetes заканчиваются на '!'. Например, create_deployment!.

      def deployment_list
        request!(:get, "/apis/extensions/v1beta1/namespaces/#{namespace}/deployments")
      end

      # Падает, если объекта нет
      def deployment(name)
        request!(:get, "/apis/extensions/v1beta1/namespaces/#{namespace}/deployments/#{name}")
      end

      # Возвращает true/false
      def deployment?(name)
        service(name)
        true
      rescue Error::NotFound
        false
      end

      def create_deployment!(spec)
        request!(:post, "/apis/extensions/v1beta1/namespaces/#{namespace}/deployments", body: spec)
      end

      def replace_deployment!(name, spec)
        request!(:put, "/apis/extensions/v1beta1/namespaces/#{namespace}/deployments/#{name}", body: spec)
      end

      def delete_deployment!(name)
        request!(:delete, "/apis/extensions/v1beta1/namespaces/#{namespace}/deployments/#{name}")
      end

      def service_list
        request!(:get, "/api/v1/namespaces/#{namespace}/services")
      end

      def service(name)
        request!(:get, "/api/v1/namespaces/#{namespace}/services/#{name}")
      end

      def service?(name)
        service(name)
        true
      rescue Error::NotFound
        false
      end

      def create_service!(spec)
        request!(:post, "/api/v1/namespaces/#{namespace}/services", body: spec)
      end

      def replace_service!(name, spec)
        request!(:put, "/api/v1/namespaces/#{namespace}/services/#{name}", body: spec)
      end

      def delete_service!(name)
        request!(:delete, "/api/v1/namespaces/#{namespace}/services/#{name}")
      end

      protected

      # query_parameters — соответствует 'Query Parameters' в документации kubernetes
      # body — hash для http-body, соответствует 'Body Parameters' в документации kubernetes, опционален
      def request!(method, path, body: nil, **query_parameters)
        with_connection do |conn|
          request_parameters = {method: method, path: path, query: query_parameters}
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
