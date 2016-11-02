module Dapp
  module Builder
    # Shell
    class Shell < Base
      [:before_install, :before_setup, :install, :setup, :build_artifact].each do |stage|
        define_method("#{stage}_checksum") do
          [dimg.config._shell.public_send("_#{stage}"),
           dimg.config._shell.public_send("_#{stage}_cache_version")].flatten
        end
        define_method("#{stage}?") { !stage_empty?(stage) }
        define_method(stage.to_s) do |image|
          image.add_command(*stage_commands(stage)) unless stage_empty?(stage)
        end
      end

      def stage_empty?(stage)
        stage_commands(stage).empty?
      end

      def stage_commands(stage)
        dimg.config._shell.public_send("_#{stage}")
      end
    end
  end
end
