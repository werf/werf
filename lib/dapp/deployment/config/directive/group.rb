module Dapp
  module Deployment
    module Config
      module Directive
        class Group < Base
          include GroupBase
          include App::InstanceMethods

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
