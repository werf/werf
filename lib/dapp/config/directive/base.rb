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
          path = path.to_s
          path = path.chomp('/') unless path == '/'
          path
        end
      end
    end
  end
end
