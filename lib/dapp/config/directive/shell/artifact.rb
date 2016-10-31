module Dapp
  module Config
    module Directive
      module Shell
        class Artifact < Dimg
          attr_reader :_build_artifact

          [:build_artifact].each do |stage|
            define_method stage do |&blk|
              unless instance_variable_get("@_#{stage}")
                instance_variable_set("@_#{stage}", StageCommand.new(&blk))
              end
              instance_variable_get("@_#{stage}")
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
        end
      end
    end
  end
end
