module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_commands") { conf[m] || [] }
        define_method(:"#{m}_do") { |image| image.build_cmd!(*send(:"#{m}_commands")) }
        define_method(:"#{m}_signature_do") { hashsum *send(:"#{m}_commands") }
      end
    end
  end
end
