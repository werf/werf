module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method :"#{m}_commands" do
          config[m]
        end

        define_method m do
          send(:"#{m}_commands")
        end

        define_method :"#{m}_key" do
          sha256([super, sha256(send(:"#{m}_commands"))])
        end
      end
    end
  end
end

