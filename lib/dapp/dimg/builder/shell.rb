module Dapp
  module Dimg
    class Builder::Shell < Builder::Base
      [:before_install, :before_setup, :install, :setup, :build_artifact].each do |stage|
        define_method("#{stage}_checksum") do
          _checksum(
            dimg.config._shell.public_send("_#{stage}_command"),
            public_send("#{stage}_version_checksum")
          )
        end
        define_method("#{stage}_version_checksum") do
          _checksum(dimg.config._shell.public_send("_#{stage}_version"), dimg.config._shell._version)
        end
        define_method("#{stage}?") { !stage_empty?(stage) }
        define_method(stage.to_s) do |image|
          image.add_command(*stage_commands(stage)) unless stage_empty?(stage)
        end
      end

      def stage_empty?(stage)
        stage_commands(stage).empty? && public_send("#{stage}_version_checksum").nil?
      end

      def stage_commands(stage)
        dimg.config._shell.public_send("_#{stage}_command")
      end
    end # Builder::Shell
  end # Dimg
end # Dapp
