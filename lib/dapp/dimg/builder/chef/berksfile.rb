module Dapp
  module Dimg
    class Builder::Chef::Berksfile
      attr_accessor :builder
      attr_reader :cookbook_path
      attr_reader :local_cookbooks

      def initialize(cookbook_path)
        @cookbook_path = Pathname.new(cookbook_path)
        @local_cookbooks = {}
      end

      def dump
        builder.send(:dump) # "friend class"
      end

      def add_local_cookbook_path(name, path)
        raise(Error::Chef, code: :berksfile_absolute_path_forbidden, data: {cookbook: name, path: path}) if path.start_with? '/'

        desc = {
          name: name,
          path: cookbook_path.join(path),
          chefignore: []
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

      class << self
        def from_file(cookbook_path:, berksfile_file_path:)
          new(cookbook_path).tap do |berksfile|
            berksfile.builder = FromFileBuilder.new(berksfile, berksfile_file_path)
          end
        end

        def from_conf(cookbook_path:, cookbooks:)
          new(cookbook_path).tap do |berksfile|
            berksfile.builder = FromConfBuilder.new(berksfile, cookbooks)
          end
        end

        protected

        def new(*args, &blk)
          super(*args, &blk)
        end
      end # << self

      class Builder
        def initialize(berksfile)
          @berksfile = berksfile
        end

        def cookbook(name, *_args, path: nil, **_kwargs)
          @berksfile.add_local_cookbook_path(name, path) if path
        end

        # rubocop:disable Style/MethodMissing
        def method_missing(*_a, &_blk)
        end
        # rubocop:enable Style/MethodMissing

        protected

        def dump
          raise
        end
      end # Builder

      class FromConfBuilder < Builder
        def initialize(berksfile, cookbooks)
          super(berksfile)

          @cookbooks = cookbooks

          @cookbooks.each do |name, desc|
            cookbook(name, desc[:version_constraint], **desc)
          end
        end

        def dump
          [].tap do |lines|
            lines << "source 'https://supermarket.chef.io'\n\n "
            @cookbooks.each do |name, desc|
              params = desc.reject {|key, _value| [:name, :version_constraint].include? key}
              lines << "cookbook #{name.inspect}, #{desc[:version_constraint].inspect}, #{params.inspect}\n"
            end
          end.join
        end
      end # FromConfBuilder

      class FromFileBuilder < Builder
        def initialize(berksfile, berksfile_file_path)
          super(berksfile)

          @berksfile_file_path = berksfile_file_path

          instance_eval(@berksfile_file_path.read, @berksfile_file_path.to_s)
        end

        def dump
          @berksfile_file_path.read
        end
      end # FileParser
    end # Builder::Chef::Berksfile
  end # Dimg
end # Dapp
