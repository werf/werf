module Dapp
  module Kube
    module Dapp
      module Command
        module SecretExtract
          def kube_secret_extract(file_path)
            secret_key_should_exist!

            data = begin
              if file_path
                raise Error::Command, code: :file_not_exist, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)

                if options[:values]
                  kube_extract_secret_values(file_path)
                else
                  kube_extract_secret_file(file_path)
                end
              else
                kube_extract_secret
              end
            end

            if (output_file_path = options[:output_file_path])
              FileUtils.mkpath File.dirname(output_file_path)
              IO.binwrite(output_file_path, data)
            else
              puts data
            end
          end

          def kube_extract_secret
            data = begin
              if $stdin.tty?
                print 'Enter secret: '
                $stdin.gets
              else
                $stdin.read
              end.to_s.chomp
            end

            if data.empty?
              exit 0
            else
              secret.extract(data)
            end
          end

          def kube_extract_secret_values(file_path)
            kube_helm_decode_json(secret, yaml_load_file(file_path)).to_yaml
          end

          def kube_extract_secret_file(file_path)
            secret.extract(IO.binread(file_path).chomp("\n"))
          end
        end
      end
    end
  end
end
