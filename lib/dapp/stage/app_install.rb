module Dapp
  module Stage
    class AppInstall < Base
      def image
        super do |image|
          builder.app_install_do(image)
        end
      end

      def signature
        hashsum [dependency_file, dependency_file_regex, builder.app_install_signature_do]
      end

      def dependency_file
        @dependency_file ||= begin
          file_path = Dir[builder.build_path('*')].detect {|x| x =~ dependency_file_regex }
          File.read(file_path) unless file_path.nil?
        end
      end

      def dependency_file?
        !dependency_file.nil?
      end

      def dependency_file_regex
        /.*\/(Gemfile|composer.json|requirement_file.txt)$/
      end
    end # AppInstall
  end # Stage
end # Dapp
