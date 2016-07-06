module Dapp
  module Config
    class Chef < Base
      attr_accessor :modules

      def initialize(main_conf, &blk)
        main_conf.builder_validation(:chef)
        @modules ||= []
        super
      end
    end
  end
end
