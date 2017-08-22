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

              loop do
                begin
                  system(kube_secret_editor, tmp_file_path.to_s)

                  encoded_data = begin
                    if options[:values]
                      encoded_values = yaml_load_file(file_path)

                      decoded_values = kube_helm_decode_json(secret, yaml_load_file(file_path))
                      updated_decoded_values = yaml_load_file(tmp_file_path)

                      deep_delete_if_key_not_exist(encoded_values, updated_decoded_values)

                      modified_decoded_values = deep_select_modified_keys(updated_decoded_values, decoded_values)
                      updated_encoded_values = encoded_values.in_depth_merge(kube_helm_encode_json(secret, modified_decoded_values))
                      deep_sort_by_same_structure(updated_encoded_values, updated_decoded_values).to_yaml
                    else
                      kube_secret_file(tmp_file_path)
                    end
                  end

                  IO.binwrite(file_path, "#{encoded_data}\n")
                  break
                rescue ::Dapp::Error::Base => e
                  log_warning(e.net_status[:data][:message])
                  print 'Do you want to change file (Y/n)?'
                  response = $stdin.noecho(&:gets).tap { print "\n" }
                  return if response.strip == 'n'
                end
              end
            end
          end

          def kube_secret_editor
            return ENV['EDITOR'] unless ENV['EDITOR'].nil?
            %w(vim vi nano).each { |editor| return editor unless shellout("which #{editor}").exitstatus.nonzero? }
            raise Error::Command, code: :editor_not_found
          end

          private

          def deep_select_modified_keys(hash1, hash2)
            {}.tap do |hash|
              hash1.each do |k, v|
                next if hash2[k] == v
                hash[k] = begin
                  if hash2[k].is_a?(Hash) && v.is_a?(Hash)
                    deep_select_modified_keys(v, hash2[k])
                  else
                    v
                  end
                end
              end
            end
          end

          def deep_delete_if_key_not_exist(hash1, hash2)
            hash1.delete_if do |k, v|
              if hash2.key?(k)
                if hash2[k].is_a?(Hash) && v.is_a?(Hash)
                  deep_delete_if_key_not_exist(v, hash2[k])
                end
                false
              else
                true
              end
            end
          end

          def deep_sort_by_same_structure(hash1, hash2)
            hash1.sort_by { |k, _| hash2.keys.index(k) }.to_h.tap do |h|
              h.select { |_, v| v.is_a?(Hash) }.each do |k, _|
                h[k] = deep_sort_by_same_structure(h[k], hash2[k])
              end
            end
          end
        end
      end
    end
  end
end
