module Dapp
  module Deployment
    module Config
      module Directive
        module Mod
          module Bootstrap
            attr_reader :_bootstrap

            def bootstrap_init_variables!
              @_bootstrap = Directive::Bootstrap.new(dapp: dapp)
            end

            def bootstrap(&blk)
              directive_eval(_bootstrap, &blk)
            end
          end
        end
      end
    end
  end
end
