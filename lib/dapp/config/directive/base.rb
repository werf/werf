module Dapp
  module Config
    module Directive
      # Base
      class Base < Config::Base
        protected

        def clone
          _clone
        end

        def clone_to_artifact
          clone
        end

        def path_format(path)
          path.to_s.chomp('/')
        end
      end
    end
  end
end
