module Dapp
  module Kube
    module Dapp
      module Command
        module SecretExtract
          def kube_secret_extract(file_path)
            secret_key_should_exist!

            if file_path.nil?
              kube_extract_secret
            else
              kube_extract_secret_file(file_path)
            end
          end

          def kube_extract_secret
            data = begin
              if $stdin.tty?
                print 'Enter secret: '
                $stdin.noecho(&:gets).tap { print "\n" }
              else
                $stdin.read
              end
            end

            unless (data = data.to_s.chomp).empty?
              puts secret.extract(data)
            end
          end

          def kube_extract_secret_file(file_path)
            raise Error::Command, code: :file_not_exist, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)

            decrypted_data = secret.extract(IO.binread(file_path).chomp("\n"))
            if (output_file_path = options[:output_file_path]).nil?
              puts decrypted_data
            else
              FileUtils.mkpath File.dirname(output_file_path)
              IO.binwrite(output_file_path, "#{decrypted_data}\n")
            end
          end
        end
      end
    end
  end
end
