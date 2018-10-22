module Dapp
  module Dimg
    module GitRepo
      class Own < Local
        def initialize(dapp)
          super(dapp, 'own', dapp.path.to_s)
        end
      end
    end
  end
end
