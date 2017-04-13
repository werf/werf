module Dapp
  module Deployment
    module Dapp
      module Command
        module Common
          def deployment
            @deployment ||= Deployment.new(dapp: self)
          end

          def secret
            @secret ||= Secret.new(ENV['DAPP_SECRET_KEY']) if ENV.key?('DAPP_SECRET_KEY')
          end
        end
      end
    end
  end
end
