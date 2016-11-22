module Dapp
  # Project
  class Project
    module SystemShellout
      SYSTEM_SHELLOUT_IMAGE = 'ubuntu:16.04'.freeze
      SYSTEM_SHELLOUT_VERSION = 4

      def system_shellout_container_name
        @system_shellout_container_name ||= begin
          "dapp_system_shellout_#{name}_#{hashsum [SYSTEM_SHELLOUT_IMAGE,
                                                   SYSTEM_SHELLOUT_VERSION,
                                                   Deps::Base::BASE_VERSION,
                                                   Deps::Gitartifact::GITARTIFACT_VERSION]}"
        end
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

      protected

      def system_shellout_container
        case system_shellout_container_state
        when :not_running
          lock(system_shellout_container_name) do
            log_secondary_process(t(code: 'process.system_shellout_container_starting'), short: true) do
              shellout!("docker rm -f #{system_shellout_container_name}")
              start_system_shellout_container!
            end
          end
        when :not_exist
          lock(system_shellout_container_name) do
            log_secondary_process(t(code: 'process.system_shellout_container_starting'), short: true) do
              start_system_shellout_container!
            end
          end
        end

        system_shellout_container_name
      end

      def system_shellout_container_state
        cmd = shellout("docker inspect -f {{.State.Running}} #{system_shellout_container_name}")
        if cmd.exitstatus.nonzero?
          :not_exist
        elsif cmd.stdout.strip == 'false'
          :not_running
        else
          :running
        end
      end

      def start_system_shellout_container!
        volumes_from = [base_container, gitartifact_container]
        project_volume = git_path ? File.dirname(git_path) : path

        # Shutdown container after 5min of inactivity
        container_loop = %(\
touch /last_access ; \
while true ; do \
  if [ ! -z "$(find /last_access -mtime 0.0035)" ] ; then \
    break ; \
  fi ; \
  sleep 1 ; \
done)

        shellout! ['docker run --detach',
                   "--name #{system_shellout_container_name}",
                   *volumes_from.map { |container| "--volumes-from #{container}" },
                   "--volume #{project_volume}:#{project_volume}",
                   SYSTEM_SHELLOUT_IMAGE,
                   "#{bash_bin} -ec '#{container_loop}'"].join(' ')

        system_shellout_livecheck!
      end

      def system_shellout_livecheck!
        # This is stupid container "live check" for now
        system_shellout! 'true'
      rescue Error::Shellout
        $stderr.puts "\033[1m\033[31mSystem-shellout container failure\033[0m"
        raise
      end

      def _to_system_shellout_command(command)
        cmd = shellout_pack ['touch /last_access', "cd #{Dir.pwd}", command].join(' && ')
        "docker exec #{system_shellout_container} #{bash_bin} -ec '#{[
          *SystemShellout.default_env_keys.map do |env_key|
            env_key = env_key.to_s.upcase
            "export #{env_key}=#{ENV[env_key]}"
          end, cmd
        ].join(' && ')}'"
      end

      class << self
        def default_env_keys
          @default_env_keys ||= []
        end
      end # << self
    end # SystemShellout
  end # Project
end # Dapp
