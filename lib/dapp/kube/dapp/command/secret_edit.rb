module Dapp
  module Kube
    module Dapp
      module Command
        module SecretEdit
          def kube_secret_edit(file_path)
            secret_key_should_exist!

            with_kube_tmp_chart_dir do
              decoded_data = begin
                raise Error::Command, code: :file_not_exist, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)

                if options[:values]
                  kube_extract_secret_values(file_path)
                else
                  kube_extract_secret_file(file_path)
                end
              end

              tmp_file_path = kube_tmp_chart_path(file_path)
              tmp_file_path.binwrite(decoded_data)
              system(kube_secret_editor, tmp_file_path.to_s)

              encoded_data = begin
                if options[:values]
                  kube_secret_values(tmp_file_path)
                else
                  kube_secret_file(tmp_file_path)
                end
              end

              IO.binwrite(file_path, "#{encoded_data}\n")
            end
          end

          def kube_secret_editor
            return ENV['EDITOR'] unless ENV['EDITOR'].nil?
            %w(vim vi nano).each { |editor| return editor unless shellout("which #{editor}").exitstatus.nonzero? }
            raise Error::Command, code: :editor_not_found
          end
        end
      end
    end
  end
end
