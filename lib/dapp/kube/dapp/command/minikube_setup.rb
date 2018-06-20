module Dapp
  module Kube
    module Dapp
      module Command
        module MinikubeSetup
          def kube_minikube_setup
            _minikube_restart_minikube

            # NOTICE: На данный момент выключено из-за непригодной
            # NOTICE: для использования медленности 'minikube mount' томов
            # _minikube_run_minikube_persisted_storage_daemon

            _minikube_wait_till_ready
            _minikube_run_registry
            _minikube_run_registry_forwarder_daemon
          end

          def _minikube_set_original_sudo_caller_process_user!
            if ENV['SUDO_UID'] and ENV['SUDO_GID']
              original_uid = ENV['SUDO_UID'].to_i
              original_gid = ENV['SUDO_GID'].to_i
              if original_uid != Process.uid && ENV['HOME'] == Etc.getpwuid(original_uid).dir
                Process::Sys.setgid(original_gid)
                Process::Sys.setuid(original_uid)
              end
            end
          end

          def _minikube_restart_minikube
            log_process("Restart minikube") do
              raise ::Dapp::Error::Command, code: :minikube_not_found if shellout('which minikube').exitstatus == 1

              Process.fork do
                _minikube_set_original_sudo_caller_process_user!

                begin
                  if shellout('minikube status').stdout.split("\n").map(&:strip).first =~ /minikube(VM)?: Running/
                    shellout! 'minikube stop', verbose: true
                  end
                  shellout! 'minikube start --insecure-registry localhost:5000', verbose: true
                rescue ::Dapp::Error::Shellout
                  exit 1
                end
              end

              _, status = Process.wait2
              raise ::Dapp::Error::Command, code: :cannot_restart_minikube unless status.success?
            end
          end

          def _minikube_run_daemon(name, &blk)
            uid = Process.euid
            if uid != 0 && File.exists?("/run/user/#{uid}")
              pid_file_path = "/run/user/#{uid}/dapp_#{name}_daemon.pid"
            elsif uid != 0
              pid_file_path = "/tmp/dapp_#{name}_daemon.pid"
            else
              pid_file_path = "/run/dapp_#{name}_daemon.pid"
            end

            old_daemon_pid = begin
              File.open(pid_file_path, 'r').read.strip.to_i
            rescue Errno::ENOENT
            end
            if old_daemon_pid
              begin
                if Process.kill(0, old_daemon_pid)
                  Process.kill('TERM', old_daemon_pid)
                end
                if Process.kill(0, old_daemon_pid)
                  Process.kill('KILL', old_daemon_pid)
                end
              rescue Errno::ESRCH
              end
            end

            daemon_pid = Process.fork do
              File.open(pid_file_path, 'w') {|f| f.write "#{Process.pid}\n"}
              yield
            end
            Process.detach(daemon_pid)

            daemon_ok = true
            begin
              Process.kill(0, daemon_pid)

              5.times do
                sleep 1
                Process.kill(0, daemon_pid)
              end
            rescue Errno::ESRCH
              daemon_ok = false
            end
            raise ::Dapp::Error::Command, code: :"#{name}_daemon_failed" unless daemon_ok
          end

          def _minikube_run_minikube_persisted_storage_daemon
            log_process("Run minikube persisted storage daemon") do
              _minikube_run_daemon(:minikube_persisted_storage) do
                _minikube_set_original_sudo_caller_process_user!

                STDIN.reopen '/dev/null'
                STDOUT.reopen '/dev/null', 'a'

                mountdir = File.join(ENV['HOME'], '.minikube', 'persisted_storage')
                FileUtils.mkdir_p mountdir

                exec "minikube mount #{mountdir}"
              end
            end
          end

          def _minikube_wait_till_ready
            log_process("Wait till minikube ready") do
              600.times do
                begin
                  return if _minikube_kubernetes.service?('kube-dns')
                rescue Kubernetes::Client::Error::ConnectionRefused
                end

                sleep 1
              end

              raise ::Dapp::Error::Command, code: :minikube_readiness_timeout
            end
          end

          def _minikube_run_registry
            log_process("Run registry") do
              _minikube_kubernetes.with_query(gracePeriodSeconds: 0) do
                if _minikube_kubernetes.replicationcontroller? _minikube_registry_replicationcontroller_spec['metadata']['name']
                  _minikube_kubernetes.delete_replicationcontroller! _minikube_registry_replicationcontroller_spec['metadata']['name']

                  shutdown_ok = false
                  600.times do
                    unless _minikube_kubernetes.replicationcontroller? _minikube_registry_replicationcontroller_spec['metadata']['name']
                      shutdown_ok = true
                      break
                    end
                    sleep 1
                  end
                  raise ::Dapp::Error::Command, code: :registry_replicationcontroller_shutdown_failed unless shutdown_ok
                end

                _minikube_kubernetes.delete_pods! labelSelector: 'k8s-app=kube-registry'
                shutdown_ok = false
                600.times do
                  unless _minikube_find_registry_pod
                    shutdown_ok = true
                    break
                  end
                  sleep 1
                end
                raise ::Dapp::Error::Command, code: :registry_pod_shutdown_failed unless shutdown_ok

                if _minikube_kubernetes.service? _minikube_registry_service_spec['metadata']['name']
                  _minikube_kubernetes.delete_service! _minikube_registry_service_spec['metadata']['name']

                  shutdown_ok = false
                  600.times do
                    unless _minikube_kubernetes.service? _minikube_registry_service_spec['metadata']['name']
                      shutdown_ok = true
                      break
                    end
                    sleep 1
                  end
                  raise ::Dapp::Error::Command, code: :registry_service_shutdown_failed unless shutdown_ok
                end

                if _minikube_kubernetes.pod? _minikube_registry_proxy_pod_spec['metadata']['name']
                  _minikube_kubernetes.delete_pod! _minikube_registry_proxy_pod_spec['metadata']['name']

                  shutdown_ok = false
                  600.times do
                    unless _minikube_kubernetes.pod? _minikube_registry_proxy_pod_spec['metadata']['name']
                      shutdown_ok = true
                      break
                    end
                    sleep 1
                  end
                  raise ::Dapp::Error::Command, code: :registry_proxy_pod_shutdown_failed unless shutdown_ok
                end

                _minikube_kubernetes.create_replicationcontroller!(_minikube_registry_replicationcontroller_spec)
                registry_pod_ok = false
                600.times do
                  if registry_pod = _minikube_find_registry_pod
                    if registry_pod['status']['phase'] == 'Running'
                      registry_pod_ok = true
                      @_minikube_registry_pod_name = registry_pod['metadata']['name']
                      break
                    end
                  end
                  sleep 1
                end
                raise ::Dapp::Error::Command, code: :registry_pod_not_ok unless registry_pod_ok

                _minikube_kubernetes.create_service!(_minikube_registry_service_spec)
                registry_service_ok = false
                600.times do
                  if _minikube_kubernetes.service? _minikube_registry_service_spec['metadata']['name']
                    registry_service_ok = true
                    break
                  end
                  sleep 1
                end
                raise ::Dapp::Error::Command, code: :registry_service_not_ok unless registry_service_ok

                _minikube_kubernetes.create_pod! _minikube_registry_proxy_pod_spec
                registry_proxy_pod_ok = false
                600.times do
                  if _minikube_kubernetes.pod? _minikube_registry_proxy_pod_spec['metadata']['name']
                    registry_proxy_pod = _minikube_kubernetes.pod(_minikube_registry_proxy_pod_spec['metadata']['name'])
                    if registry_proxy_pod['status']['phase'] == 'Running'
                      registry_proxy_pod_ok = true
                      break
                    end
                  end
                  sleep 1
                end
                raise ::Dapp::Error::Command, code: :registry_proxy_pod_not_ok unless registry_proxy_pod_ok
              end
            end
          end

          def _minikube_run_registry_forwarder_daemon
            log_process("Run registry forwarder daemon") do
              registry_port_in_use = begin
                Socket.tcp('localhost', 5000).close
                true
              rescue Errno::ECONNREFUSED
                false
              end
              raise ::Dapp::Error::Command, code: :registry_port_in_use, data: {host: 'localhost', port: 5000} if registry_port_in_use

              _minikube_run_daemon(:registry_forwarder) do
                STDIN.reopen '/dev/null'
                STDOUT.reopen '/dev/null', 'a'

                exec "kubectl port-forward --namespace kube-system #{_minikube_registry_pod_name} 5000:5000"
              end
            end
          end

          def _minikube_find_registry_pod
            _minikube_kubernetes.pod_list(labelSelector: "k8s-app=kube-registry")['items'].first
          end

          def _minikube_registry_pod_name
            @_minikube_registry_pod_name
          end

          def _minikube_kubernetes
            @_minikube_kubernetes ||= Kubernetes::Client.new(namespace: 'kube-system')
          end

          def _minikube_registry_replicationcontroller_spec
            {
              'metadata' => {
                'name' => 'kube-registry',
                'namespace' => 'kube-system',
                'labels' => {
                  'k8s-app' => 'kube-registry',
                  'version' => 'v0',
                  'kubernetes.io/cluster-service' => 'true'
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
                      'version' => 'v0',
                      'kubernetes.io/cluster-service' => 'true'
                    }
                  },
                  'spec' => {
                    'containers' => [
                      {
                        'name' => 'registry',
                        'image' => 'registry:2',
                        'env' => [
                          {
                            'name' => 'REGISTRY_HTTP_ADDR',
                            'value' => ':5000'
                          },
                          {
                            'name' => 'REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY',
                            'value' => '/var/lib/registry'
                          },
                          {
                            'name' => 'REGISTRY_STORAGE_DELETE_ENABLED',
                            'value' => 'true',
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
                        'hostPath' => {
                          'path' => '/var/lib/docker/registry_data'
                        }
                      }
                    ]
                  }
                }
              }
            }
          end

          def _minikube_registry_service_spec
            {
              'metadata' => {
                'name' => 'kube-registry',
                'namespace' => 'kube-system',
                'labels' => {
                  'k8s-app' => 'kube-registry',
                  'kubernetes.io/cluster-service' => 'true',
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

          def _minikube_registry_proxy_pod_spec
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
        end
      end
    end
  end
end
