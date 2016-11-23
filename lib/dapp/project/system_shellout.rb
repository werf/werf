module Dapp
  # Project
  class Project
    module SystemShellout
      SYSTEM_SHELLOUT_IMAGE = 'ubuntu:16.04'.freeze

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

      def system_shellout_extra(volume: [], &blk)
        old = _system_shellout_extra_opts.dup
        _system_shellout_extra_opts[:volume] ||= []
        _system_shellout_extra_opts[:volume] += Array(volume)

        yield if block_given?
      ensure
        @_system_shellout_extra_opts = old
      end

      protected

      def _to_system_shellout_command(command)
        volumes_from = [base_container, gitartifact_container]
        project_volume = git_path ? File.dirname(git_path) : path

        ['docker run --rm',
         *volumes_from.map { |container| "--volumes-from #{container}" },
         *Array(_system_shellout_extra_opts[:volume]).map { |volume| "--volume #{volume}:#{volume}" },
         "--volume #{project_volume}:#{project_volume}",
         *SystemShellout.default_env_keys.map { |env_key|
           env_key = env_key.to_s.upcase
           "--env #{env_key}=#{ENV[env_key]}"
         },
         SYSTEM_SHELLOUT_IMAGE,
         "#{bash_bin} -ec '#{shellout_pack(command)}'"
        ].join(' ')
      end

      def _system_shellout_extra_opts
        @_system_shellout_extra_opts ||= {}
      end

      class << self
        def default_env_keys
          @default_env_keys ||= []
        end
      end # << self
    end # SystemShellout
  end # Project
end # Dapp
