module Dapp
  module Builder
    # Shell
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_checksum") do
          [application.config._shell.public_send("_#{m}"),
           application.config._shell.public_send("_#{m}_cache_version")].flatten
        end
        define_method(:"#{m}") do |image|
          image.add_commands(*application.config._shell.public_send("_#{m}"))
        end
      end
    end
  end
end
