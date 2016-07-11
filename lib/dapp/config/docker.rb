module Dapp
  module Config
    class Docker < Base
      attr_accessor :from, :exposes

      def initialize(main_conf, &blk)
        @exposes = []
        super
      end

      def expose(*args)
        exposes.push(*args.flatten)
      end
    end
  end
end
