module Dapp
  module Config
    # Docker
    class Docker
      attr_reader :_expose
      attr_reader :_from_cache_version

      def initialize
        @_expose = []
      end

      def from(image_name, cache_version: nil)
        @_from = image_name
        @_from_cache_version = cache_version
      end

      def expose(*args)
        @_expose.concat(args)
      end

      def _from
        @_from || raise(Error::Config, code: :docker_from_is_not_defined)
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
