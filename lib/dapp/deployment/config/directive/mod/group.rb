module Dapp
  module Deployment
    module Config
      module Directive
        module Mod
          module Group
            def _app
              (@_app + @_group.map(&:_app)).flatten
            end

            protected

            def deploy_init_variables!
              @_group = []
              @_app   = []
            end
          end
        end
      end
    end
  end
end
