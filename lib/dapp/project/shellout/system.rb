module Dapp
  # Project
  class Project
    module Shellout
      # System
      module System
        SYSTEM_SHELLOUT_IMAGE = 'ubuntu:14.04'.freeze
        SYSTEM_SHELLOUT_VERSION = 3

        def system_shellout_container_name
          "dapp_system_shellout_#{hashsum [SYSTEM_SHELLOUT_VERSION,
                                           SYSTEM_SHELLOUT_IMAGE,
                                           Deps::Base::BASE_VERSION,
                                           Deps::Gitartifact::GITARTIFACT_VERSION]}"
        end

        def system_shellout_container
          do_livecheck = false

          @system_shellout_container ||= begin
            lock(system_shellout_container_name) do
              cmd = shellout("docker inspect -f {{.State.Running}} #{system_shellout_container_name}")
              if cmd.exitstatus.nonzero?
                start_container = true
              elsif cmd.stdout.strip == 'false'
                shellout!("docker rm -f #{system_shellout_container_name}")
                start_container = true
              else
                start_container = false
              end

              if start_container
                volumes_from = [base_container, gitartifact_container]
                log_secondary_process(t(code: 'process.system_shellout_container_loading'), short: true) do
                  shellout! ['docker run --detach --privileged',
                             "--name #{system_shellout_container_name}",
                             *volumes_from.map { |container| "--volumes-from #{container}" },
                             '--volume /:/.system_shellout_root',
                             "#{SYSTEM_SHELLOUT_IMAGE} #{bash_path} -ec 'while true ; do sleep 1 ; done'"].join(' ')

                  shellout! ["docker exec #{system_shellout_container_name}",
                             "bash -ec '#{[
                               'mkdir -p /.system_shellout_root/.dapp',
                               'mount --rbind /.dapp /.system_shellout_root/.dapp'
                             ].join(' && ')}'"].join(' ')
                end
              end
            end

            do_livecheck = true
            system_shellout_container_name
          end

          system_shellout_livecheck! if do_livecheck

          @system_shellout_container
        end

        def system_shellout(command, raise_error: false, **kwargs)
          if raise_error
            shellout! _to_system_shellout_command(command), **kwargs
          else
            shellout _to_system_shellout_command(command), **kwargs
          end
        end

        def system_shellout!(command, **kwargs)
          system_shellout(command, raise_error: true, **kwargs)
        end

        private

        def system_shellout_livecheck!
          # This is stupid container "live check" for now
          system_shellout! 'true'
        rescue Error::Shellout
          $stderr.puts "\033[1m\033[31mSystem shellout container failure, " +
                       "try to remove if error persists: " +
                       "docker rm -f #{system_shellout_container_name}\033[0m"
          raise
        end

        def _to_system_shellout_command(command)
          cmd = shellout_pack ["cd #{Dir.pwd}", command].join(' && ')
          "docker exec #{system_shellout_container} chroot /.system_shellout_root #{bash_path} -ec '#{[
            *System.default_env_keys.map do |env_key|
              env_key = env_key.to_s.upcase
              "export #{env_key}=#{ENV[env_key]}"
            end, cmd
          ].join(' && ')}'"
        end

        public

        class << self
          def default_env_keys
            @default_env_keys ||= []
          end
        end # << self
      end # System
    end
  end # Project
end # Dapp
