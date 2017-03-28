module Dapp
  module Dimg
    class Builder::Chef::CookbookMetadata
      attr_accessor :builder
      attr_accessor :name
      attr_accessor :version

      def depends
        @depends ||= {}
      end

      def dump
        builder.send(:dump) # "friend class"
      end

      class << self
        def from_file(metadata_file_path:)
          new.tap do |metadata|
            metadata.builder = FromFileBuilder.new(metadata, metadata_file_path)
          end
        end

        def from_conf(name:, version:, cookbooks:)
          new.tap do |metadata|
            metadata.builder = FromConfBuilder.new(metadata, name, version, cookbooks)
          end
        end

        protected

        def new(*args, &blk)
          super(*args, &blk)
        end
      end # << self

      class Builder
        def initialize(metadata)
          @metadata = metadata
        end

        def name(name)
          @metadata.name = name
        end

        def version(version)
          @metadata.version = version
        end

        def depends(dependency, version_constraint = nil, **kwargs, &_blk)
          @metadata.depends[dependency] = {}.tap do |desc|
            desc.update(kwargs)
            desc[:dependency] = dependency
            desc[:version_constraint] = version_constraint if version_constraint
          end
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

      class FromFileBuilder < Builder
        def initialize(metadata, metadata_file_path)
          super(metadata)

          @metadata_file_path = metadata_file_path

          instance_eval(@metadata_file_path.read, @metadata_file_path.to_s)
        end

        def dump
          @metadata_file_path.read
        end
      end # FromFileBuilder

      class FromConfBuilder < Builder
        def initialize(metadata, name, version, cookbooks)
          super(metadata)

          @cookbooks = cookbooks

          @cookbooks.each do |cname, desc|
            depends(cname, desc[:version_constraint], **desc)
          end

          self.name name
          self.version version
        end

        def dump
          [].tap do |lines|
            lines << "name #{@metadata.name.inspect}\n"
            lines << "version #{@metadata.version.inspect}\n"

            @cookbooks.keys.each do |cookbook|
              lines << "depends #{cookbook.inspect}\n" unless cookbook.start_with? 'dimod-'
            end
          end.join
        end
      end # FromConfBuilder
    end # Builder::Chef::CookbookMetadata
  end # Dimg
end # Dapp
