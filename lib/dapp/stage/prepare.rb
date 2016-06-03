module Dapp
  module Stage
    class Prepare < Base
      include Dapp::Stage::Mod::Centos7
      include Dapp::Stage::Mod::Ubuntu1404
      include Dapp::Stage::Mod::Ubuntu1604

      def image
        super do |image|
          send(_image_method, image)
          image.build_options[:expose] = conf[:exposes] unless conf[:exposes].nil?
        end
      end

      def _image_method
        from = from_image_name
        :"from_#{from.to_s.split(/[:.]/).join}".tap do |from_method|
          raise "unsupported docker image '#{from}'" unless respond_to?(from_method)
        end
      end

      def from_image_name
        builder.conf[:from]
      end

      def signature
        image.signature
      end
    end # Prepare
  end # Builder
end # Dapp
