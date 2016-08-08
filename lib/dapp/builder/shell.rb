module Dapp
  module Builder
    # Shell
    class Shell < Base
      [:infra_install, :infra_setup, :install, :setup].each do |stage|
        define_method(:"#{stage}_checksum") do
          [application.config._shell.public_send("_#{stage}"),
           application.config._shell.public_send("_#{stage}_cache_version")].flatten
        end
        define_method(:"#{stage}") do |image|
          commands = application.config._shell.public_send("_#{stage}")
          image.add_commands("export DAPP_BUILD_STAGE=#{stage}", *commands) if commands
        end
      end
    end
  end
end
