module Dapp
  module Kube
    module Dapp
      module Command
        module SecretFileEncrypt
          def kube_secret_file_encrypt(file_path)
            raise Error::Command, code: :secret_key_not_found if secret.nil?
            raise Error::Command, code: :file_not_exist, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)

            encrypted_data = secret.generate(IO.binread(file_path))
            if (output_file_path = options[:output_file_path]).nil?
              puts encrypted_data
            else
              FileUtils.mkpath File.dirname(output_file_path)
              IO.binwrite(output_file_path, "#{encrypted_data}\n")
            end
          end
        end
      end
    end
  end
end
