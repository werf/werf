module Dapp
  module Deployment
    module Dapp
      module Command
        module MinikubeSetup
          def deployment_minikube_setup
            _minikube_start_minikube
            _minikube_create_environment
            _minikube_run_forwarder
          end

          def _minikube_start_minikube
            raise Error::Command, code: :minikube_not_found if shellout('which minikube').exitstatus == 1
            log_process(:start_minikube) do
              Process.fork do
                uid = ENV['SUDO_UID'].to_i
                gid = ENV['SUDO_GID'].to_i
                if uid != Process.uid && ENV['HOME'] == Etc.getpwuid(uid).dir
                  Process::Sys.setgid(gid)
                  Process::Sys.setuid(uid)
                end
                exec 'minikube start --insecure-registry localhost:5000'
              end
              _, status = Process.wait2
              raise Error::Command, code: :minikube_not_started unless status.success? && _minikube_started?
            end
          end

          def _minikube_started?
            10.times do
              begin
                return true if _minikube_kubernetes.service?('kube-dns')
              rescue Excon::Error::Socket
                sleep 1
              end
            end
            false
          end

          def _minikube_create_environment
            log_process(:create_environment) do
              _minikube_kubernetes.with_query(gracePeriodSeconds: 0) do
                [:replicationcontroller, :service, :pod].each do |object|
                  with_log_indent do
                    _minikube_create_or_replace_kubernetes_object(object)
                  end
                end
              end
            end
          end

          def _minikube_create_or_replace_kubernetes_object(object)
            spec = public_send("_minikube_#{object}_spec")
            name = spec['metadata']['name']

            if _minikube_kubernetes.send("#{object}?", name)
              log_secondary_process(:"delete #{object}", short: true, quiet: log_verbose?) do
                _minikube_kubernetes.send("delete_#{object}!", name)
                loop do
                  break unless _minikube_kubernetes.send("#{object}?", name)
                  sleep(1)
                end
              end
            end

            log_secondary_process(:"create #{object}", short: true, quiet: log_verbose?) do
              _minikube_kubernetes.send("create_#{object}!", spec)
            end
          end

          def _minikube_replicationcontroller_spec
            {
              'metadata' => {
                'name' => 'kube-registry-v0',
                'namespace' => 'kube-system',
                'labels' => {
                  'k8s-app' => 'kube-registry',
                  'version' => 'v0'
                }
              },
              'spec' => {
                'replicas' => 1,
                'selector' => {
                  'k8s-app' => 'kube-registry',
                  'version' => 'v0'
                },
                'template' => {
                  'metadata' => {
                    'labels' => {
                      'k8s-app' => 'kube-registry',
                      'version' => 'v0'
                    }
                  },
                  'spec' => {
                    'containers' => [
                      {
                        'name' => 'registry',
                        'image' => 'registry:2',
                        'resources' => {
                          'limits' => {
                            'cpu' => '100m',
                            'memory' => '100Mi'
                          }
                        },
                        'env' => [
                          {
                            'name' => 'REGISTRY_HTTP_ADDR',
                            'value' => ':5000'
                          },
                          {
                            'name' => 'REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY',
                            'value' => '/var/lib/registry'
                          }
                        ],
                        'volumeMounts' => [
                          {
                            'name' => 'image-store',
                            'mountPath' => '/var/lib/registry'
                          }
                        ],
                        'ports' => [
                          {
                            'containerPort' => 5000,
                            'name' => 'registry',
                            'protocol' => 'TCP'
                          }
                        ]
                      }
                    ],
                    'volumes' => [
                      {
                        'name' => 'image-store',
                        'emptyDir' => {}
                      }
                    ]
                  }
                }
              }
            }
          end

          def _minikube_service_spec
            {
              'metadata' => {
                'name' => 'kube-registry',
                'namespace' => 'kube-system',
                'labels' => {
                  'k8s-app' => 'kube-registry',
                  'kubernetes.io/name' => 'KubeRegistry'
                }
              },
              'spec' => {
                'selector' => {
                  'k8s-app' => 'kube-registry'
                },
                'ports' => [
                  {
                    'name' => 'registry',
                    'port' => 5000,
                    'protocol' => 'TCP'
                  }
                ]
              }
            }
          end

          def _minikube_pod_spec
            {
              'metadata' => {
                'name' => 'kube-registry-proxy',
                'namespace' => 'kube-system'
              },
              'spec' => {
                'containers' => [
                  {
                    'name' => 'kube-registry-proxy',
                    'image' => 'gcr.io/google_containers/kube-registry-proxy:0.3',
                    'resources' => {
                      'limits' => {
                        'cpu' => '100m',
                        'memory' => '50Mi'
                      }
                    },
                    'env' => [
                      {
                        'name' => 'REGISTRY_HOST',
                        'value' => 'kube-registry.kube-system.svc.cluster.local'
                      },
                      {
                        'name' => 'REGISTRY_PORT',
                        'value' => '5000'
                      },
                      {
                        'name' => 'FORWARD_PORT',
                        'value' => '5000'
                      }
                    ],
                    'ports' => [
                      {
                        'name' => 'registry',
                        'containerPort' => 5000,
                        'hostPort' => 5000
                      }
                    ]
                  }
                ]
              }
            }
          end

          def _minikube_run_forwarder
            pod_name = _minikube_pod_name

            log_process(:run_forwarder, short: true) do
              Process.fork do
                STDIN.reopen '/dev/null'
                STDOUT.reopen '/dev/null', 'a'
                STDERR.reopen '/dev/null', 'a'

                exec "kubectl port-forward --namespace kube-system #{pod_name} 5000:5000"
              end
            end
          end

          def _minikube_pod_name
            raise unless (pods = _minikube_kubernetes.pod_list(labelSelector: "k8s-app=kube-registry")['items']).one?
            pods.first['metadata']['name']
          end

          def _minikube_kubernetes
            @kubernetes ||= Kubernetes.new(namespace: 'kube-system')
          end
        end
      end
    end
  end
end
