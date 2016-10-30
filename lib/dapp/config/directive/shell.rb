module Dapp
  module Config
    module Directive
      class Shell < Base
        attr_reader :_version
        attr_reader :_before_install, :_before_setup, :_install, :_setup

        def version(value)
          @_version = value
        end

        def before_install(&blk)
          @_before_install ||= StageCommand.new(project: _project, &blk)
        end

        def before_setup(&blk)
          @_before_setup ||= StageCommand.new(project: _project, &blk)
        end

        def install(&blk)
          @_install ||= StageCommand.new(project: _project, &blk)
        end

        def setup(&blk)
          @_setup ||= StageCommand.new(project: _project, &blk)
        end

        protected

        class StageCommand < Directive::Base
          attr_reader :_version
          attr_reader :_command

          def initialize(project:)
            @_command = []

            super
          end

          def command(*args)
            @_before_setup.concat(args)
          end

          def version(value)
            @_version = value
          end
        end
      end
    end
  end
end
