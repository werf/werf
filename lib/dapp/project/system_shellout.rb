module Dapp
  # Project
  class Project
    module SystemShellout
      SYSTEM_SHELLOUT_IMAGE = 'ubuntu:16.04'.freeze

      def system_shellout(command, raise_error: false, **kwargs)
        system_shellout_extra(volume: (git_path ? File.dirname(git_path) : path)) do
          begin
            if raise_error
              shellout! _to_system_shellout_command(command), **kwargs
            else
              shellout _to_system_shellout_command(command), **kwargs
            end
          rescue Error::Shellout
            log_warning(
              desc: { code: :launched_command,
                      data: { command: _to_system_shellout_command(command) },
                      context: :system_shellout },
              quiet: !log_verbose?
            )
            raise
          end
        end
      end

      def system_shellout!(command, **kwargs)
        system_shellout(command, raise_error: true, **kwargs)
      end

      def system_shellout_extra(volume: [], workdir: nil, &blk)
        old = system_shellout_opts.dup

        system_shellout_opts[:volume] += Array(volume)
        system_shellout_opts[:workdir] = workdir if workdir

        yield if block_given?
      ensure
        @system_shellout_opts = old
      end

      protected

      def system_shellout_opts
        @system_shellout_opts ||= {volume: []}
      end

      def _to_system_shellout_command(command)
        volumes_from = [base_container, gitartifact_container]

        ['docker run',
         '--rm',
         "--workdir #{system_shellout_opts[:workdir] || Dir.pwd}",
         *volumes_from.map { |container| "--volumes-from #{container}" },
         *Array(system_shellout_opts[:volume]).map { |volume| "--volume #{volume}:#{volume}" },
         *SystemShellout.default_env_keys.map { |env_key|
           env_key = env_key.to_s.upcase
           "--env #{env_key}=#{ENV[env_key]}"
         },
         SYSTEM_SHELLOUT_IMAGE,
         "#{bash_bin} -ec '#{shellout_pack(command)}'"
        ].join(' ')
      end

      class << self
        def default_env_keys
          @default_env_keys ||= []
        end
      end # << self
    end # SystemShellout
  end # Project
end # Dapp
