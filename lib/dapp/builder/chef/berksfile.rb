module Dapp
  module Builder
    class Chef < Base
      # Berksfile
      class Berksfile
        # Parser
        class Parser
          def initialize(berksfile)
            @berksfile = berksfile
            parse
          end

          def cookbook(name, path: nil, **_kwargs)
            @berksfile.add_local_cookbook_path(name, path) if path
          end

          def method_missing(*_a, &blk)
          end

          private

          def parse
            instance_eval(@berksfile.path.read, @berksfile.path.to_s)
          end
        end # Parser

        attr_reader :path
        attr_reader :local_cookbooks

        def initialize(home_path, path)
          @home_path = home_path
          @path = path
          @local_cookbooks = {}
          @parser = Parser.new(self)
        end

        def add_local_cookbook_path(name, path)
          @local_cookbooks[name] = {
            name: name,
            path: @home_path.join(path)
          }
        end

        def local_cookbook?(name)
          local_cookbooks.key? name
        end

        def local_cookbook(name)
          local_cookbooks[name]
        end
      end # Berksfile
    end # Chef
  end # Builder
end # Dapp
