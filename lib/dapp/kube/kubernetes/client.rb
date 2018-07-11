module Dapp
  module Kube
    module Kubernetes
      # TODO endpoints can be gathered from api-server by api discovery.
      K8S_API_ENDPOINTS = {
        '1.6' => {
          '/api/v1' => [:service, :replicationcontroller, :pod, :podtemplate, ],
          '/apis/apps/v1beta1' => [:deployment, :statefulset, ],
          '/apis/extensions/v1beta1' => [:replicaset, :daemonset, ],
          '/apis/batch/v1' => [:job, ],
          '/apis/batch/v2aplha1' => [:cronjob, ],
        },
        '1.7' => {
          '/api/v1' => [:service, :replicationcontroller, :pod, :podtemplate, ],
          '/apis/apps/v1beta1' => [:deployment, :statefulset, ],
          '/apis/extensions/v1beta1' => [:replicaset, :daemonset, ],
          '/apis/batch/v1' => [:job, ],
          '/apis/batch/v2aplha1' => [:cronjob, ],
        },
        '1.8' => {
          '/api/v1' => [:service, :replicationcontroller, :pod, :podtemplate, ],
          '/apis/apps/v1beta2' => [:daemonset, :deployment, :replicaset, :statefulset, ],
          '/apis/batch/v1' => [:job, ],
          '/apis/batch/v1beta1' => [:cronjob, ],
        },
        '1.9' => {
          '/api/v1' => [:service, :replicationcontroller, :pod, :podtemplate, ],
          '/apis/apps/v1' => [:daemonset, :deployment, :replicaset, :statefulset, ],
          '/apis/batch/v1' => [:job, ],
          '/apis/batch/v1beta1' => [:cronjob, ],
        },
        '1.10' => {
          '/api/v1' => [:service, :replicationcontroller, :pod, :podtemplate, ],
          '/apis/apps/v1' => [:daemonset, :deployment, :replicaset, :statefulset, ],
          '/apis/batch/v1' => [:job, ],
          '/apis/batch/v1beta1' => [:cronjob, ],
        },
        'stable' => {
          '/api/v1' => [:service, :replicationcontroller, :pod, :podtemplate, ],
          '/apis/batch/v1' => [:job, ],
        },
      }

      class Client
        include Helper::YAML
        extend Helper::YAML

        ::Dapp::Dapp::Shellout::Base.default_env_keys << 'KUBECONFIG'

        attr_reader :config
        attr_reader :context
        attr_reader :namespace
        attr_reader :timeout

        def initialize(config, context, namespace, timeout: nil)
          @config = config
          @context = context
          @namespace = namespace
          @timeout = timeout
          @query_parameters = {}
          @cluster_version
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
        # В каждом методе происходит выбор api на основе версии кластера

        def resource_endpoint_path(resource)
          K8S_API_ENDPOINTS[cluster_version()].map do |path, resources|
            resources.include?(resource) ? path : nil
          end.compact.first
        end

        [
          :service, :replicationcontroller, :pod, :podtemplate,
          :daemonset, :deployment, :replicaset, :statefulset,
          :job,
          :cronjob,
        ].each do |resource|
          define_method :"#{resource}_list" do |**query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:get, "#{api_path}/namespaces/#{namespace}/#{resource}s", **query_parameters)
          end

          define_method resource do |name, **query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:get, "#{api_path}/namespaces/#{namespace}/#{resource}s/#{name}", **query_parameters)
          end

          define_method "#{resource}_status" do |name, **query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:get, "#{api_path}/namespaces/#{namespace}/#{resource}s/#{name}/status", **query_parameters)
          end

          define_method :"create_#{resource}!" do |spec, **query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:post, "#{api_path}/namespaces/#{namespace}/#{resource}s", body: spec, **query_parameters)
          end

          define_method :"replace_#{resource}!" do |name, spec, **query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:put, "#{api_path}/namespaces/#{namespace}/#{resource}s/#{name}", body: spec, **query_parameters)
          end

          define_method :"delete_#{resource}!" do |name, **query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:delete, "#{api_path}/namespaces/#{namespace}/#{resource}s/#{name}", **query_parameters)
          end

          define_method :"delete_#{resource}s!" do |**query_parameters|
            api_path = resource_endpoint_path(resource)
            request!(:delete, "#{api_path}/namespaces/#{namespace}/#{resource}s", **query_parameters)
          end

          define_method :"#{resource}?" do |name, **query_parameters|
            public_send(:"#{resource}_list", **query_parameters)['items'].map { |item| item['metadata']['name'] }.include?(name)
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

        # minikube returns empty major and minor. Fallback to stable only apis for minikube setup
        def cluster_version(**query_parameters)
          version_obj = request!(:get, "/version", **query_parameters)
          @cluster_version ||= begin
            major = version_obj['major']
            minor = version_obj['minor'].gsub(/\+$/,'')
            k8s_version = "#{version_obj['major']}.#{version_obj['minor']}"
            if K8S_API_ENDPOINTS.has_key?(k8s_version)
              k8s_version
            else
              "stable"
            end
          end
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
        # excon_parameters — соответствует connection-опциям Excon
        # body — hash для http-body, соответствует 'Body Parameters' в документации kubernetes, опционален
        def request!(method, path, body: nil, excon_parameters: {}, response_body_parameters: {}, **query_parameters)
          if method == :get
            excon_parameters[:idempotent] = true
            excon_parameters[:retry_limit] = 6
            excon_parameters[:retry_interval] = 5
          end

          excon_parameters[:connect_timeout] ||= timeout
          excon_parameters[:read_timeout] ||= timeout
          excon_parameters[:write_timeout] ||= timeout

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
            err_data[:response_raw_body] = response.body
            if response_body = (JSON.parse(response.body) rescue nil)
              err_data[:response_body] = response_body
            end
            err_data[:request_parameters] = request_parameters

            if response.status.to_s.start_with? '5'
              raise Error::Default, code: :server_error, data: err_data
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
          connection = begin
            context_config = config.context_config(context)
            cluster_config = config.cluster_config(context_config['cluster'])
            Excon.new(cluster_config['server'], **kube_server_options(excon_parameters)).tap(&:get)
          rescue Excon::Error::Socket => err
            raise Error::ConnectionRefused,
                  code: :server_connection_refused,
                  data: { url: cluster_config['server'], error: err.message }
          end

          return yield connection
        end

        def kube_server_options(excon_parameters = {})
          {}.tap do |opts|
            context_config = config.context_config(context)
            user_config = config.user_config(context_config['user'])
            cluster_config = config.cluster_config(context_config['cluster'])

            client_cert = user_config['client-certificate']
            opts[:client_cert] = client_cert if client_cert

            client_cert_data= user_config['client-certificate-data']
            opts[:client_cert_data] = Base64.decode64(client_cert_data) if client_cert_data

            client_key = user_config['client-key']
            opts[:client_key] = client_key if client_key

            client_key_data = user_config['client-key-data']
            opts[:client_key_data] = Base64.decode64(client_key_data) if client_key_data

            ssl_cert_store = OpenSSL::X509::Store.new
            if ssl_ca_file = cluster_config['certificate-authority']
              ssl_cert_store.add_file ssl_ca_file
            elsif ssl_ca_data = cluster_config['certificate-authority-data']
              ssl_cert_store.add_cert OpenSSL::X509::Certificate.new(Base64.decode64(ssl_ca_data))
            end
            opts[:ssl_cert_store] = ssl_cert_store

            opts[:ssl_ca_file] = nil

            opts[:middlewares] = [*Excon.defaults[:middlewares], Excon::Middleware::RedirectFollower]

            opts.merge!(excon_parameters)
          end
        end
      end # Client
    end # Kubernetes
  end # Kube
end # Dapp
