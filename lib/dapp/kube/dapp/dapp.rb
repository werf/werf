module Dapp
  module Kube
    module Dapp
      module Dapp
        include Command::SecretGenerate
        include Command::SecretKeyGenerate
        include Command::SecretFileEncrypt
        include Command::Common
      end
    end
  end
end

::Dapp::Dapp.send(:include, ::Dapp::Kube::Dapp::Dapp)
