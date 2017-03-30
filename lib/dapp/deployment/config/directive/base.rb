module Dapp
  module Deployment
    module Config
      module Directive
        class Base < ::Dapp::Config::Directive::Base
          def hostname_pattern
            '^[a-z0-9_-]*[a-z0-9]$'
          end
        end
      end
    end
  end
end
