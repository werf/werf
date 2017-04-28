module Dapp
  module Deployment
    module Dapp
      module Command
        module Common
          def deployment
            @deployment ||= Deployment.new(config: config, dapp: self)
          end
        end
      end
    end
  end
end
