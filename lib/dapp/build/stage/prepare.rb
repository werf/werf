module Dapp
  module Build
    module Stage
      class Prepare < Base
        include Mod::Centos7
        include Mod::Ubuntu1404
        include Mod::Ubuntu1604
        # centos7_image
        # centos7_signature

        def signature
          # FIXME send(:"#{build.conf[:from].to_s.split(/[:.]/).join}_signature")
          image.signature
        end

        def image
          # FIXME DO NOT USE SUPER!
          super do |image|
            send(image_constructor_method, image)
          end
        end

        protected

        # FIXME remove
        def from_image
          DockerImage.new from: build.conf[:from]
        end

        private

        def image_constructor_method
          :"from_#{build.conf[:from].to_s.split(/[:.]/).join}"
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
