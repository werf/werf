module Dapp
  module Config
    # Docker
    class Docker
      attr_reader :_volume, :_expose, :_env, :_label, :_cmd, :_onbuild, :_workdir, :_user, :_entrypoint
      attr_reader :_from_cache_version

      def initialize
        @_volume = []
        @_expose = []
        @_env = {}
        @_label = {}
        @_cmd = []
        @_onbuild = []
      end

      def from(image, cache_version: nil)
        raise(Error::Config, code: :docker_from_incorrect, data: { name: image }) unless image.match(/^[[^ ].]+:[[^ ].]+$/)
        @_from = image
        @_from_cache_version = cache_version
      end

      def volume(*args)
        @_volume.concat(args)
      end

      def expose(*args)
        @_expose.concat(args)
      end

      def env(**options)
        @_env.merge!(options)
      end

      def label(**options)
        @_label.merge!(options)
      end

      def cmd(*args)
        @_cmd.concat(args)
      end

      def onbuild(*args)
        @_onbuild.concat(args)
      end

      def workdir(val)
        @_workdir = val
      end

      def user(val)
        @_user = val
      end

      def entrypoint(*cmd_with_args)
        @_entrypoint = cmd_with_args.flatten
      end

      def _from
        @_from || raise(Error::Config, code: :docker_from_not_defined)
      end

      def _change_options
        {
          volume: _volume,
          expose: _expose,
          env: _env,
          label: _label,
          cmd: _cmd,
          onbuild: _onbuild,
          workdir: _workdir,
          user: _user,
          entrypoint: _entrypoint
        }
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
