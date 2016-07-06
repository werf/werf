module Dapp
  module Config
    class Docker < Base
      attr_accessor :from, :exposes

      # FIXME docker.expose 80, 90, 100

      def initialize(main_conf, &blk)
        @exposes ||= []
        super
      end
    end
  end
end
