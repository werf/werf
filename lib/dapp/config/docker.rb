module Dapp
  module Config
    class Docker
      attr_reader :_from
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
        @_expose.push(*args.flatten)
      end

      def to_h
        {
          from:               _from,
          from_cache_version: _from_cache_version,
          expose:             _expose
        }.select { |_k, v| !v.nil? and !v.empty? }
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
