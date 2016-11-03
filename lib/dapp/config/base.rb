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

      def marshal_dup(obj)
        Marshal.load(Marshal.dump(obj))
      end
    end
  end
end
