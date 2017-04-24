module Dapp
  module Deployment
    module Config
      module Directive
        class Job < Base
          attr_reader :_run, :_dimg

          def initialize(*args)
            super
            @_run = []
          end

          def run(*args)
            sub_directive_eval { @_run.concat(args.flatten) }
          end

          def dimg(name)
            sub_directive_eval { @_dimg = name }
          end

          def empty?
            _run.empty? && _dimg.nil?
          end
        end
      end
    end
  end
end
