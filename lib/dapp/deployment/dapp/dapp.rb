module Dapp
  module Deployment
    module Dapp
      module Dapp
        include Command::Apply
        include Command::SecretKeyGenerate
        include Command::SecretGenerate
        include Command::Mrproper
        include Command::Common
      end
    end
  end
end

::Dapp::Dapp.send(:include, ::Dapp::Deployment::Dapp::Dapp)
