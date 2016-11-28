module Dapp
  module Config
    module Directive
      # Base
      class Base < Config::Base
        def initialize(project:, &blk)
          @project = project

          instance_eval(&blk) unless blk.nil?
        end

        protected

        attr_reader :project

        def clone
          marshal_dup(self)
        end

        def clone_to_artifact
          clone
        end
      end
    end
  end
end
