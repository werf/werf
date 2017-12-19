module Dapp
  module Helper
    module YAML
      def yaml_load_file(file_path, hash: true, default: {})
        return default if (context = File.read(file_path).strip).empty?
        yaml_load(context, hash: hash)
      rescue ::Dapp::Error::Dapp => e
        raise ::Dapp::Error::Dapp, code: :yaml_file_incorrect, data: { file: file_path, message: e.net_status[:data][:message] }
      end

      def yaml_load(string, hash: true)
        ::YAML::load(string).tap do |res|
          if hash && !res.is_a?(Hash)
            raise ::Dapp::Error::Dapp,
                  code: :yaml_incorrect,
                  data: { message: "unexpected yaml data \n>>> START YAML\n#{string.strip}\n>>> END YAML\n" }
          end
        end
      rescue Psych::SyntaxError => e
        raise ::Dapp::Error::Dapp, code: :yaml_incorrect, data: { message: e.message }
      end
    end
  end # Helper
end # Dapp
