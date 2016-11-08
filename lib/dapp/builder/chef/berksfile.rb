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

          def cookbook(name, *_args, path: nil, **_kwargs)
            @berksfile.add_local_cookbook_path(name, path) if path
          end

          # rubocop:disable Style/MethodMissing
          def method_missing(*_a, &_blk)
          end
          # rubocop:enable Style/MethodMissing

          private

          def parse
            instance_eval(@berksfile.path.read, @berksfile.path.to_s)
          end
        end # Parser

        attr_reader :home_path
        attr_reader :path
        attr_reader :local_cookbooks

        def initialize(home_path, path)
          @home_path = home_path
          @path = path
          @local_cookbooks = {}
          @parser = Parser.new(self)
        end

        def add_local_cookbook_path(name, path)
          raise(::Dapp::Builder::Chef::Error, code: :berksfile_absolute_path_forbidden,
                                              data: { cookbook: name, path: path }) if path.start_with? '/'

          desc = {
            name: name,
            path: home_path.join(path),
            chefignore: [],
          }

          if desc[:path].join('chefignore').exist?
            chefignore_patterns = desc[:path].join('chefignore').read.split("\n").map(&:strip)
            desc[:chefignore] = Dir[*chefignore_patterns.map {|pattern| desc[:path].join(pattern)}]
              .map(&Pathname.method(:new))
          end

          @local_cookbooks[name] = desc
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
