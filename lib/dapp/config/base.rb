module Dapp
  module Config
    class Base
      def initialize(project:, &blk)
        @project = project
        instance_eval(&blk) unless blk.nil?
      end

      protected

      attr_reader :project

      def marshal_dup(obj)
        Marshal.load(Marshal.dump(obj))
      end
    end
  end
end
