module Dapp
  # Application
  class Application
    # SystemShellout
    module SystemShellout
      SYSTEM_SHELLOUT_IMAGE = 'ubuntu:14.04'
      SYSTEM_SHELLOUT_VERSION = 2

      def system_shellout_container_name
        "dapp_system_shellout_#{hashsum [SYSTEM_SHELLOUT_VERSION,
                                         SYSTEM_SHELLOUT_IMAGE,
                                         Deps::Gitartifact::GITARTIFACT_IMAGE]}"
      end

      def system_shellout_container
        @system_shellout_container ||= begin
          if shellout("docker inspect #{system_shellout_container_name}").exitstatus.nonzero?
            volumes_from = [gitartifact_container]
            log_secondary_process(t(code: 'process.system_shellout_container_loading'), short: true) do
              shellout! ['docker run --detach --privileged',
                         "--name #{system_shellout_container_name}",
                         *volumes_from.map { |container| "--volumes-from #{container}" },
                         '--volume /:/.system_shellout_root',
                         "#{SYSTEM_SHELLOUT_IMAGE} bash -ec 'while true ; do sleep 1 ; done'"].join(' ')

              shellout! ["docker exec #{system_shellout_container_name}",
                         "bash -ec '#{[
                           'mkdir -p /.system_shellout_root/.dapp',
                           'mount --rbind /.dapp /.system_shellout_root/.dapp',
                           'mount --rbind /usr/bin /.system_shellout_root/usr/bin',
                           'apt-get update -qq',
                           'apt-get install -qq openssh-client',
                         ].join(' && ')}'"].join(' ')
            end
          end

          system_shellout_container_name
        end
      end

      def system_shellout(command, **kwargs)
        shellout _to_system_shellout_command(command), **kwargs
      end

      def system_shellout!(command, **kwargs)
        shellout! _to_system_shellout_command(command), **kwargs
      end

      private

      def _to_system_shellout_command(command)
        cmd = shellout_pack ["cd #{Dir.pwd}", command].join(' && ')
        "docker exec #{system_shellout_container} chroot /.system_shellout_root bash -ec '#{cmd}'"
      end
    end # SystemShellout
  end # Application
end # Dapp
