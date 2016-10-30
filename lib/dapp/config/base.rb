module Dapp
  module Config
    class Base
      attr_reader :_project

      def initialize(project:, &blk)
        @_project = project
        instance_eval(&blk) unless blk.nil?
      end
    end
  end
end
