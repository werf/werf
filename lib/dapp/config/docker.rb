module Dapp
  module Config
    class Docker < Base
      attr_reader :_from
      attr_reader :_expose

      def initialize
        @_expose = []
        super
      end

      def from(image_name, cache_version: nil)
        @_from = image_name
        cache_version(from: cache_version) unless cache_version.nil?
      end

      def expose(*args)
        @_expose.push(*args.flatten)
      end

      def to_h
        {
          from:    _from,
          exposes: _exposes
        }
      end
    end
  end
end
