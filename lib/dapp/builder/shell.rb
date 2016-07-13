module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_checksum") { [application.config._shell.public_send("_#{m}"),
                                           application.config._shell.public_send("_#{m}_cache_version")].flatten }
        define_method(:"#{m}") { |image| image.add_commands(*application.config._shell.public_send("_#{m}")) }
      end
    end
  end
end
