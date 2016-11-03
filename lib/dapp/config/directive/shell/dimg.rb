module Dapp
  module Config
    module Directive
      module Shell
        class Dimg < Directive::Base
          attr_reader :_version
          attr_reader :_before_install, :_before_setup, :_install, :_setup

          def version(value)
            @_version = value
          end

          protected

          def empty?
            (_before_install_command + _before_setup_command + _install_command + _setup_command).empty?
          end

          class StageCommand < Directive::Base
            attr_reader :_version
            attr_reader :_command

            def initialize
              @_command = []
              super
            end

            def command(*args)
              @_command.concat(args)
            end

            def version(value)
              @_version = value
            end
          end

          def self.stage_command_generator(stage)
            define_method stage do |&blk|
              (variable = instance_variable_get("@_#{stage}") || StageCommand.new).instance_eval(&blk)
              instance_variable_set("@_#{stage}", variable)
            end

            define_method "_#{stage}_command" do
              return [] if (variable = instance_variable_get("@_#{stage}")).nil?
              variable._command
            end

            define_method "_#{stage}_version" do
              return [] if (variable = instance_variable_get("@_#{stage}")).nil?
              variable._version || _version
            end
          end
          [:before_install, :before_setup, :install, :setup].each(&method(:stage_command_generator))

          protected

          def clone_to_artifact
            Artifact.new.tap do |shell|
              instance_variables.each do |variable|
                shell.instance_variable_set(variable, instance_variable_get(marshal_dup(variable)))
              end
            end
          end
        end
      end
    end
  end
end
