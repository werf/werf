module Dapp
  module Build
    module Stage
      class Prepare < Base
        include Mod::Centos7
        include Mod::Ubuntu1404
        include Mod::Ubuntu1604

        def signature
          image.signature
        end

        protected

        def image
          super do |image|
            send(image_constructor_method, image)
          end
        end

        def from_image_name
          build.conf[:from]
        end

        private

        def image_constructor_method
          :"from_#{from_image_name.to_s.split(/[:.]/).join}"
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
