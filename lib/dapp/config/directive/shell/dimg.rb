module Dapp
  module Config
    module Directive
      module Shell
        # Dimg
        class Dimg < Directive::Base
          attr_reader :_version
          attr_reader :_before_install, :_before_setup, :_install, :_setup

          def version(value)
            @_version = value
          end

          def self.stage_command_generator(stage)
            define_method stage do |&blk|
              (variable = instance_variable_get("@_#{stage}") || StageCommand.new(project: project)).instance_eval(&blk)
              instance_variable_set("@_#{stage}", variable)
            end

            define_method "_#{stage}_command" do
              return [] if (variable = instance_variable_get("@_#{stage}")).nil?
              variable._run
            end

            define_method "_#{stage}_version" do
              return [] if (variable = instance_variable_get("@_#{stage}")).nil?
              variable._version || _version
            end
          end
          [:before_install, :before_setup, :install, :setup].each(&method(:stage_command_generator))

          # StageCommand
          class StageCommand < Directive::Base
            attr_reader :_version
            attr_reader :_run

            def initialize(**kwargs, &blk)
              @_run = []

              super(**kwargs, &blk)
            end

            def run(*args)
              @_run.concat(args)
            end

            def version(value)
              @_version = value
            end
          end

          protected

          def empty?
            (_before_install_command + _before_setup_command + _install_command + _setup_command).empty?
          end
        end
      end
    end
  end
end
