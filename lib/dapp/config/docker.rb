module Dapp
  module Config
    class Docker < Base
      attr_accessor :from, :exposes

      def initialize(main_conf, &blk)
        @exposes ||= []
        super
      end
    end
  end
end
