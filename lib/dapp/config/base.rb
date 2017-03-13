module Dapp
  module Config
    class Base
      def initialize(dapp:, &blk)
        @dapp = dapp
        instance_eval(&blk) if block_given?
      end

      protected

      attr_reader :dapp

      def ref_variables
        [:@dapp]
      end

      def marshal_dump
        instance_variables
          .reject {|variable| ref_variables.include? variable}
          .map {|variable| [variable, instance_variable_get(variable)]}
      end

      def marshal_load(variable_values)
        variable_values.each do |variable, value|
          instance_variable_set(variable, value)
        end

        self
      end

      def _clone
        Marshal.load(Marshal.dump(self)).tap do |obj|
          _set_ref_variables_to(obj)
        end
      end

      def _clone_to(obj)
        obj.marshal_load(marshal_dump)
        _set_ref_variables_to(obj)
      end

      def _set_ref_variables_to(obj)
        ref_variables.each do |ref_variable|
          obj.instance_variable_set(ref_variable, instance_variable_get(ref_variable))
        end
      end
    end
  end
end
