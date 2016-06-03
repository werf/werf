module Dapp
  module Stage
    class Base
      include CommonHelper

      attr_accessor :prev, :next

      def build
        return if image_exist?
        prev.build if prev
        build_image!
      end
    end # Base
  end # Stage
end # Dapp