module Dapp
  module Builder
    class Shell < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_commands") { conf[m] }
        define_method(m) do
          cmds = send(:"#{m}_commands") || []
          Image.new(from: send(:"#{m}_from")).tap { |image| image.build_cmd!(*cmds) }
        end
        define_method(:"#{m}_key") { sha256([super(), send(:"#{m}_commands")]) }
      end

      def app_install_key
        sha256([super(), app_install_commands])
      end

      def app_setup_key
        sha256([super(), app_setup_commands])
      end
    end
  end
end

