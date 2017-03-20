module Dapp
  module Config
    module Directive
      class Base < Config::Base
        def clone
          _clone
        end

        protected

        def sub_directive_eval
          yield if block_given?
          self
        end

        def path_format(path)
          path.to_s.chomp('/')
        end
      end
    end
  end
end
