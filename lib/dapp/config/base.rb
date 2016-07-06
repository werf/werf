module Dapp
  module Config
    class Base
      include Dapp::CommonHelper

      def initialize(main_conf = nil, **options)
        @cache_version = {}
        @main_conf = main_conf

        options.each do |k, v|
          if respond_to? k
            send(:"#{k}=", v)
          else
            raise "Object '#{self.class.to_s}' doesn't have attribute '#{k}'!"
          end
        end

        yield self if block_given?
      end

      def cache_version(key = nil)
        @cache_version[key]
      end

      def cache_version=(value = nil, **options)
        if value.nil?
          options.each { |k, v| @cache_version[k] = v }
        else
          @cache_version[nil] = value
        end
      end

      protected

      attr_reader :main_conf
    end
  end
end
