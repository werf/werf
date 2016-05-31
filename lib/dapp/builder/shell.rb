module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_commands") { conf[m] || [] }
        define_method(m) { Image.new(from: send(:"#{m}_from")).tap { |image| image.build_cmd!(*send(:"#{m}_commands")) } }
        define_method(:"#{m}_key") { hashsum [super(), send(:"#{m}_commands")] }
      end
    end
  end
end

