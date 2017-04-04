module Dapp
  module Deployment
    module Dapp
      module Dapp
        include Command::Apply
        include Command::Mrproper
      end
    end
  end
end

::Dapp::Dapp.send(:include, ::Dapp::Deployment::Dapp::Dapp)
