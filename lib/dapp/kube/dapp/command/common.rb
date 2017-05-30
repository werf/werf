module Dapp
  module Kube
    module Dapp
      module Command
        module Common
          def secret
            @secret ||= Secret.new(ENV['DAPP_SECRET_KEY']) if ENV.key?('DAPP_SECRET_KEY')
          end
        end
      end
    end
  end
end
