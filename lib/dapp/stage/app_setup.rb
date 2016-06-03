module Dapp
  module Stage
    class AppSetup < Base
      def image
        super do |image|
          builder.app_setup_do(image)
        end
      end

      def signature
        hashsum [app_setup_file, builder.app_setup_signature_do]
      end

      def app_setup_file
        @app_setup_file ||= begin
          File.read(app_setup_file_path) if app_setup_file?
        end
      end

      def app_setup_file?
        File.exist?(app_setup_file_path)
      end

      def app_setup_file_path
        builder.build_path('.app_setup')
      end
    end # AppSetup
  end # Stage
end # Dapp
