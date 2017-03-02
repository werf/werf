module Dapp
  module Dimg::Builder
    class Chef::CookbookMetadata
      class << self
        def from_file(metadata_file_path)
          metadata = new
          FileParser.new(metadata_file_path, metadata)
          metadata
        end

        def from_conf(conf)
          # TODO
        end
      end # << self

      class FileParser
        def initialize(metadata_file_path, metadata)
          @metadata_file_path = metadata_file_path
          @metadata = metadata

          parse
        end

        def name(name)
          @metadata.name = name
        end

        def version(version)
          @metadata.version = version
        end

        def depends(dependency, *_a, &_blk)
          @metadata.depends << dependency.to_s
        end

        # rubocop:disable Style/MethodMissing
        def method_missing(*_a, &_blk)
        end
        # rubocop:enable Style/MethodMissing

        private

        def parse
          instance_eval(@metadata_file_path.read, @metadata_file_path.to_s)
        end
      end # FileParser

      attr_accessor :name
      attr_accessor :version

      def depends
        @depends ||= []
      end
    end # Chef::CookbookMetadata
  end # Dimg::Builder
end # Dapp
