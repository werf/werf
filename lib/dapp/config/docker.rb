module Dapp
  module Config
    # Docker
    class Docker
      attr_reader :_expose, :_workdir, :_env
      attr_reader :_from_cache_version

      def initialize
        @_expose = []
        @_env = []
      end

      def from(image_name, cache_version: nil)
        @_from = image_name
        @_from_cache_version = cache_version
      end

      def expose(*args)
        @_expose.concat(args)
      end

      def workdir(path)
        @_workdir = path
      end

      def env(*args)
        @_env.concat(args)
      end

      def _from
        @_from || fail(Error::Config, code: :docker_from_not_defined)
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
