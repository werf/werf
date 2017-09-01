module Dapp
  module Kube
    module Dapp
      module Command
        module SecretGenerate
          def kube_secret_generate(file_path)
            secret_key_should_exist!

            data = begin
              if file_path
                kube_secret_file_validate!(file_path)

                if options[:values]
                  kube_secret_values(file_path)
                else
                  kube_secret_file(file_path)
                end
              else
                kube_secret
              end
            end

            if (output_file_path = options[:output_file_path])
              FileUtils.mkpath File.dirname(output_file_path)
              IO.binwrite(output_file_path, "#{data}\n")
            else
              puts data
            end
          end

          def kube_secret
            data = begin
              if $stdin.tty?
                print 'Enter secret: '
                $stdin.noecho(&:gets).tap { print "\n" }
              else
                $stdin.read
              end.to_s.chomp
            end

            if data.empty?
              exit 0
            else
              secret.generate(data)
            end
          end

          def kube_secret_values(file_path)
            kube_helm_encode_json(secret, yaml_load_file(file_path)).to_yaml
          end

          def kube_secret_file(file_path)
            secret.generate(IO.binread(file_path))
          end
        end
      end
    end
  end
end
