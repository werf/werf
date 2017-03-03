module Dapp
  module Config
    class Base
      def initialize(dapp:, &blk)
        @dapp = dapp
        instance_eval(&blk) if block_given?
      end

      protected

      attr_reader :dapp

      def marshal_dump
        instance_variables
          .reject {|variable| variable == :@dapp}
          .map {|variable| [variable, instance_variable_get(variable)]}
      end

      def marshal_load(variable_values)
        variable_values.each do |variable, value|
          instance_variable_set(variable, value)
        end

        self
      end

      def _clone
        Marshal.load Marshal.dump(self)
      end

      def _clone_to(obj)
        obj.marshal_load marshal_dump
      end
    end
  end
end
