module Dapp
  module Builder
    # Shell
    class Shell < Base
      [:infra_install, :infra_setup, :install, :setup].each do |stage|
        define_method("#{stage}_checksum") do
          [application.config._shell.public_send("_#{stage}"),
           application.config._shell.public_send("_#{stage}_cache_version")].flatten
        end
        define_method("#{stage}?") { !stage_empty?(stage) }
        define_method("#{stage}") do |image|
          image.add_command("export DAPP_BUILD_STAGE=#{stage}",
                            *stage_commands(stage)) unless stage_empty?(stage)
        end
      end

      def stage_empty?(stage)
        stage_commands(stage).empty?
      end

      def stage_commands(stage)
        application.config._shell.public_send("_#{stage}")
      end
    end
  end
end
