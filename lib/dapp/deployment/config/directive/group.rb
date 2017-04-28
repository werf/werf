module Dapp
  module Deployment
    module Config
      module Directive
        class Group < Base
          include Mod::Group
          include App::InstanceMethods

          def group(&blk)
            Group.new(dapp: dapp).tap do |group|
              group.instance_eval(&blk) if block_given?
              @_group << group
            end
          end

          def app(name = nil, &blk)
            App.new(name, dapp: dapp).tap do |app|
              pass_to(app)
              app.instance_eval(&blk) if block_given?
              @_app << app
            end
          end
        end
      end
    end
  end
end
