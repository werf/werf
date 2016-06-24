module Dapp
  module Build
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_checksum") { conf[m] || [] }
        define_method(:"#{m}_do") { |image| image.add_commands(*conf[m]) }
      end
    end
  end
end
