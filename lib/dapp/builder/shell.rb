module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_checksum") { application.conf[m] || [] }
        define_method(:"#{m}") { |image| image.add_commands(*application.conf[m]) }
      end
    end
  end
end
