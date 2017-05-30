module Dapp
  module Deployment
    module Dapp
      module Dapp
        include Command::Apply
        include Command::MinikubeSetup
        include Command::Mrproper
        include Command::Common

        include Dappfile
      end
    end
  end
end

::Dapp::Dapp.send(:include, ::Dapp::Deployment::Dapp::Dapp)
