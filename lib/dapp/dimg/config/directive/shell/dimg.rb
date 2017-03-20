module Dapp
  module Dimg
    module Config
      module Directive
        module Shell
          class Dimg < Directive::Base
            attr_reader :_version
            attr_reader :_before_install, :_before_setup, :_install, :_setup

            def version(value)
              sub_directive_eval { @_version = value }
            end

            def self.stage_command_generator(stage)
              define_method stage do |&blk|
                (instance_variable_get("@_#{stage}") || StageCommand.new(dapp: dapp)).tap do |variable|
                  instance_variable_set("@_#{stage}", directive_eval(variable, &blk))
                end
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

            def empty?
              (_before_install_command + _before_setup_command + _install_command + _setup_command).empty?
            end

            def clone_to_artifact
              _clone_to Artifact.new(dapp: dapp)
            end

            class StageCommand < Directive::Base
              attr_reader :_version
              attr_reader :_run

              def initialize(**kwargs, &blk)
                @_run = []
                super(**kwargs, &blk)
              end

              def run(*args)
                sub_directive_eval { @_run.concat(args) }
              end

              def version(value)
                sub_directive_eval { @_version = value }
              end
            end
          end
        end
      end
    end
  end
end
