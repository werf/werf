module Dapp
  module Stage
    class Base
      include CommonHelper

      attr_accessor :prev, :next
      attr_reader :builder

      def initialize(builder)
        @builder = builder
      end

      def build
        return if image_exist?
        prev.build if prev
        build_image!
      end

      def image_exist?
        docker.image_exist? image_name
      end

      def build_image!
        docker.build_image! image: image, name: image_name
      end

      def from_image_name
        @from_image_name || (prev.image_name if prev) || begin
          raise "missing from_image_name"
        end
      end

      def signature
        raise
      end

      def image_name
        "dapp:#{signature}"
      end

      def image
        @image ||= begin
          Image.new(from: from_image_name).tap do |image|
            yield image if block_given?
          end
        end
      end
    end # Base
  end # Stage
end # Dapp
