module Dapp
  module Dimg
    module Builder
      class Chef::Berksfile
        class << self
          def from_file(cookbook_path, berksfile_file_path)
            berksfile = self.new(cookbook_path)
            FileParser.new(berksfile_file_path, berksfile)
            berksfile
          end

          def from_conf(cookbook_path, conf)
            # TODO
          end
        end # << self

        class FileParser
          def initialize(berksfile_file_path, berksfile)
            @berksfile_file_path = berksfile_file_path
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
            instance_eval(@berksfile_file_path.read, @berksfile_file_path.to_s)
          end
        end # FileParser

        attr_reader :cookbook_path
        attr_reader :local_cookbooks

        def initialize(cookbook_path)
          @cookbook_path = Pathname.new(cookbook_path)
          @local_cookbooks = {}
        end

        def dump
          #FIXME
        end

        def add_local_cookbook_path(name, path)
          raise(::Dapp::Dimg::Builder::Chef::Error, code: :berksfile_absolute_path_forbidden,
                                                    data: { cookbook: name, path: path }) if path.start_with? '/'

          desc = {
            name: name,
            path: cookbook_path.join(path),
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
      end # Chef::Berksfile
    end # Builder
  end # Dimg
end # Dapp
