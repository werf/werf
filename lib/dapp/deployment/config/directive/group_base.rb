module Dapp
  module Deployment
    module Config
      module Directive
        module GroupBase
          def group(&blk)
            Group.new(dapp: dapp).tap do |group|
              group.instance_eval(&blk) if block_given?
              @_group << group
            end
          end

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
