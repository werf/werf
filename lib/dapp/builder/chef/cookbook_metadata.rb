module Dapp
  module Builder
    class Chef < Base
      # CookbookMetadata
      class CookbookMetadata
        # Parser
        class Parser
          def initialize(cookbook_metadata)
            @cookbook_metadata = cookbook_metadata
            parse
          end

          def name(name)
            @cookbook_metadata.name = name
          end

          def version(version)
            @cookbook_metadata.version = version
          end

          private

          def parse
            instance_eval(@cookbook_metadata.path.read, @cookbook_metadata.path.to_s)
          end
        end # Parser

        attr_reader :path
        attr_accessor :name
        attr_accessor :version

        def initialize(path)
          @path = path
          @parser = Parser.new(self)
        end
      end # CookbookMetadata
    end # Chef
  end # Builder
end # Dapp
