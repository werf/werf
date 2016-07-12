module Dapp
  module Config
    class Base
      attr_reader :_cache_version

      def initialize
        @_cache_version = {}
        yield self if block_given?
      end

      def cache_version(value = nil, **options)
        if value.nil?
          options.each { |k, v| @_cache_version[k] = v }
        else
          @_cache_version[nil] = value
        end
      end

      def _cache_version(key = nil)
        @_cache_version[key]
      end

      protected

      def clone_with_marshal
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
