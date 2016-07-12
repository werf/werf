module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_checksum") { [application.conf._shell.public_send("_#{m}"), application.conf._shell._cache_version(m)].flatten }
        define_method(:"#{m}") { |image| image.add_commands(*application.conf._shell.public_send("_#{m}")) }
      end
    end
  end
end
