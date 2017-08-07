module Dapp
  module Config
    class Config < Directive::Base
      def dev_mode
        @_dev_mode = true
      end

      def _dev_mode
        !!@_dev_mode
      end

      def after_parsing!
        do_all!('_after_parsing!')
      end

      def validate!
        do_all!('_validate!')
      end
    end
  end
end
