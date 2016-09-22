module Dapp
  module Config
    module Directive
      # TmpDir
      class TmpDir
        attr_reader :_store

        def initialize
          @_store = []
        end

        def store(*args)
          _store.concat(args)
        end

        def unstore(*args)
          @_store -= args
        end

        def reset
          @_store = []
        end

        protected

        def empty?
          _store.empty?
        end

        def clone
          Marshal.load(Marshal.dump(self))
        end
      end
    end
  end
end
