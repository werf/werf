module Dapp
  module Config
    module Directive
      # Base
      class Base < Config::Base
        def initialize(&blk)
          instance_eval(&blk) unless blk.nil?
        end

        protected

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
