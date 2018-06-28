module Dapp
  class Dapp
    module Command
      module Example
        module List
          def example_list
            _example_list.each(&method(:puts))
          end
        end
      end
    end
  end
end
