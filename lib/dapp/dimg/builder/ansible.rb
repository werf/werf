module Dapp
  module Dimg
    class Builder::Ansible < Builder::Base
      ANSIBLE_IMAGE_VERSION = "2.4.1.0-1"

      def ansible_bin
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/bin/ansible"
      end

      def ansible_image
        "dappdeps/ansible:#{ANSIBLE_IMAGE_VERSION}"
      end

      def ansible_container_name
        "dappdeps_ansible_#{ANSIBLE_IMAGE_VERSION}"
      end

      def ansible_container
        @ansible_container ||= begin
          is_container_exist = proc{dimg.dapp.shellout("#{dimg.dapp.host_docker} inspect #{ansible_container_name}").exitstatus.zero?}
          if !is_container_exist.call
            dimg.dapp.lock("dappdeps.container.#{ansible_container_name}", default_timeout: 600) do
              if !is_container_exist.call
                dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.ansible_container_creating', data: {name: ansible_container_name}), short: true) do
                  dimg.dapp.shellout!(
                    ["#{dimg.dapp.host_docker} create",
                     "--name #{ansible_container_name}",
                     "--volume /.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION} #{ansible_image}"].join(' ')
                  )
                end
              end
            end
          end
          ansible_container_name
        end
      end
        
      def ansible_config
        # DAPP_LOAD_CONFIG_PATH=https://github.com/flant/dapp/new/dappfile-yml-ansible/playground/test1/ansible-conf.yml
        dimg.config._ansible
      end
    end # Builder::Ansible
  end # Dimg
end # Dapp
