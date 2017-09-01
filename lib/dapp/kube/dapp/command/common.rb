module Dapp
  module Kube
    module Dapp
      module Command
        module Common
          def kube_check_helm!
            raise Error::Command, code: :helm_not_found if shellout('which helm').exitstatus == 1
          end

          def kube_release_name
            "#{name}-#{kube_namespace}"
          end

          def kube_namespace
            kubernetes.namespace
          end

          { encode: :generate, decode: :extract }.each do |type, secret_method|
            define_method "kube_helm_#{type}_json" do |secret, json|
              change_json_value = proc do |value|
                case value
                when Array then value.map { |v| change_json_value.call(v) }
                when Hash then send(:"kube_helm_#{type}_json", secret, value)
                when '' then ''
                else
                  secret.nil? ? '' : secret.public_send(secret_method, value)
                end

                json.each { |k, v| json[k] = change_json_value.call(v) }
              end
            end
          end

          def kube_secret_file_validate!(file_path)
            raise Error::Command, code: :secret_file_not_found, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)
            raise Error::Command, code: :secret_file_empty, data: { path: File.expand_path(file_path) } if File.read(file_path).strip.empty?
          end

          def secret_key_should_exist!
            raise(Error::Command,
              code: :secret_key_not_found,
              data: {not_found_in: secret_key_not_found_in.join(', ')}
            ) if secret.nil?
          end

          def kube_chart_secret_path(*path)
            kube_chart_path(kube_chart_secret_dir_name, *path).expand_path
          end

          def kube_chart_secret_dir_name
            'secret'
          end

          def kube_chart_path(*path)
            self.path('.helm', *path).expand_path
          end

          def with_kube_tmp_chart_dir
            yield if block_given?
          ensure
            FileUtils.rm_rf(kube_tmp_chart_path)
          end

          def kube_tmp_chart_path(*path)
            @kube_tmp_path ||= Dir.mktmpdir('dapp-helm-chart-', tmp_base_dir)
            make_path(@kube_tmp_path, *path).expand_path.tap { |p| p.parent.mkpath }
          end

          def secret
            @secret ||= begin
              unless (secret_key = ENV['DAPP_SECRET_KEY'])
                secret_key_not_found_in << '`DAPP_SECRET_KEY`'

                if dappfile_exists?
                  file_path = path('.dapp_secret_key')
                  if file_path.file?
                    secret_key = path('.dapp_secret_key').read.chomp
                  else
                    secret_key_not_found_in << "`#{file_path}`"
                  end
                else
                  log_warning(desc: { code: :secret_key_dappfile_not_found })
                end
              end

              Secret.new(secret_key) if secret_key
            end
          end

          def secret_key_not_found_in
            @secret_key_not_found_in ||= []
          end

          def kubernetes
            @kubernetes ||= begin
              namespace = options[:namespace].nil? ? nil : options[:namespace].tr('_', '-')
              Kubernetes::Client.new(namespace: namespace)
            end
          end
        end
      end
    end
  end
end
