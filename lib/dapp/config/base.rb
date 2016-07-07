module Dapp
  module Config
    class Base
      include Dapp::CommonHelper

      def initialize(main_conf = nil, **options)
        @main_conf = main_conf

        options.each do |k, v|
          if respond_to? k
            send(:"#{k}=", v)
          else
            raise "Object '#{object_name}' doesn't have attribute '#{k}'!"
          end
        end

        yield self if block_given?
      end

      protected

      attr_reader :main_conf

      private

      def object_name
        self.class.to_s.split('::').last
      end
    end
  end
end
