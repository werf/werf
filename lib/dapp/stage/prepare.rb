module Dapp
  module Stage
    class Prepare < Base
      include Dapp::Stage::Mod::Centos7
      include Dapp::Stage::Mod::Ubuntu1404
      include Dapp::Stage::Mod::Ubuntu1604
    end # Prepare
  end # Builder
end # Dapp
