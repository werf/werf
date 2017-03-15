module Dapp
  module Dimg
    module Config
      module Directive
        class Base < ::Dapp::Config::Directive::Base
          def clone_to_artifact
            clone
          end
        end
      end
    end
  end
end
