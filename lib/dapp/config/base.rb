module Dapp
  module Config
    # Base
    class Base
      def initialize(project:, &blk)
        @project = project
        instance_eval(&blk) if block_given?
      end

      protected

      attr_reader :project

      def marshal_dump
        instance_variables
          .reject {|variable| variable == :@project}
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
    end
  end
end
