module Dapp
  module Kube
    module Dapp
      module Dapp
        include Helper::YAML

        include Command::Deploy
        include Command::Dismiss
        include Command::SecretKeyGenerate
        include Command::SecretGenerate
        include Command::SecretExtract
        include Command::SecretRegenerate
        include Command::SecretEdit
        include Command::MinikubeSetup
        include Command::ChartCreate
        include Command::Render
        include Command::Lint
        include Command::ValueGet
        include Command::Common
        include Command::Ruby2Go
      end
    end
  end
end

::Dapp::Dapp.send(:include, ::Dapp::Kube::Dapp::Dapp)
