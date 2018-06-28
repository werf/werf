module Dapp
  class Dapp
    module Command
      module Sample
        module List
          def sample_list
            _sample_list.each(&method(:puts))
          end
        end
      end
    end
  end
end
