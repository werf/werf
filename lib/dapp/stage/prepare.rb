module Dapp
  module Stage
    class Prepare < Base
      include Dapp::Stage::Mod::Centos7
      include Dapp::Stage::Mod::Ubuntu1404
      include Dapp::Stage::Mod::Ubuntu1604

      def image
        super do |image|
          image.build_cmd! 'ololo', 'trololo'
        end
      end
    end # Prepare
  end # Builder
end # Dapp
