module Dapp
  module Kube
    module Dapp
      module Command
        module SecretGenerate
          def kube_secret_generate(file_path)
            secret_key_should_exist!

            if file_path.nil?
              kube_secret
            else
              kube_secret_file(file_path)
            end
          end

          def kube_secret
            data = begin
              if $stdin.tty?
                print 'Enter secret: '
                $stdin.noecho(&:gets).tap { print "\n" }
              else
                $stdin.read
              end
            end

            unless (data = data.to_s.chomp).empty?
              puts secret.generate(data)
            end
          end

          def kube_secret_file(file_path)
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
