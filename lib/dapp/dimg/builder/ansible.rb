module Dapp
  module Dimg
    class Builder::Ansible < Builder::Base
      ANSIBLE_IMAGE_VERSION = "2.4.4.0-2"

      def ansible_bin
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/bin/ansible"
      end

      def ansible_playbook_bin
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/bin/ansible-playbook"
      end

      def python_path
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/bin/python"
      end

      def ansible_playbook_solo_cmd
        "#{ansible_playbook_bin} -c local"
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

      def stage_config(stage)
        # DAPP_LOAD_CONFIG_PATH=https://github.com/flant/dapp/new/dappfile-yml-ansible/playground/test1/ansible-conf.yml
        dimg.config._ansible[stage.to_s].map do |dapp_task|
          {}.tap do |task|
            task.merge!(dapp_task['config'])
            task['tags'] = [].tap do |tags|
              tags << dapp_task['config']['tags']
              tags << stage.to_s
              tags << dapp_task['dump_config_section']
            end.flatten.compact
          end
        end || []
      end

      def doc_config
        dimg.config._ansible['dump_config_doc']
      end

      def stage_playbook(stage)
        @stage_playbooks ||= {}
        @stage_playbooks[stage.to_s] ||= begin
          [{
            'hosts' => 'all',
            'gather_facts' => 'no',
            'tasks' => [],
           }].tap do |playbook|
            playbook[0]['tasks'].concat(stage_config(stage))
          end
        end
      end

      def install_playbook(stage)
        @install_playbooks ||= {}
        @install_playbooks[stage.to_s] ||= true.tap do
          stage_tmp_path(stage).join('playbook.yml').write YAML.dump(stage_playbook(stage))
          # generate inventory with localhost and python in dappdeps-ansible
          stage_tmp_path(stage).join('hosts').write %{
localhost ansible_raw_live_stdout=yes ansible_script_live_stdout=yes ansible_python_interpreter=#{python_path}
}
          # generate ansible config for solo mode
          stage_tmp_path(stage).join('ansible.cfg').write %{
[defaults]
inventory = #{container_playbook_path}/hosts
transport = local
; do not generate retry files in ro volumes
retry_files_enabled = False
; more verbose stdout like ad-hoc ansible command from flant/ansible fork
stdout_callback = live
; force color
force_color = 1
[privilege_escalation]
become = yes
become_method = sudo
become_exe = #{dimg.dapp.sudo_bin}
become_flags = -E
}

         # save config dump for pretty errors
         stage_tmp_path(stage).join('dapp-config.yml').write dimg.config._ansible['dump_config_doc']
        end
      end

      def container_playbook_yaml_path(stage)
        install_playbook(stage)
        container_playbook_path.join('playbook.yml')
      end

      %i(before_install install before_setup setup build_artifact).each do |stage|
        define_method("#{stage}?") {
          !stage_empty?(stage)
        }

        define_method("#{stage}_checksum") do
          dimg.hashsum [JSON.dump(stage_config(stage))]
        end

        define_method(stage.to_s) do |image|
          unless stage_empty?(stage)
            image.add_env('ANSIBLE_CONFIG', container_playbook_path.join('ansible.cfg'))
            image.add_env('DAPP_DUMP_CONFIG_DOC_PATH', container_playbook_path.join('dapp-config.yml'))
            image.add_volumes_from("#{ansible_container}:ro")
            image.add_volume "#{stage_tmp_path(stage)}:#{container_playbook_path}:ro"
            image.add_command [ansible_playbook_bin,
                               container_playbook_yaml_path(stage)].join(' ')
          end
        end
      end

      def before_build_check
      end

      def before_dimg_should_be_built_check
      end

      def stage_empty?(stage)
        stage_config(stage).empty?
      end

      # host directory in tmp_dir
      def stage_tmp_path(stage)
        dimg.tmp_path(dimg.config._name, "ansible-playbook-#{stage.to_s}").tap {|p| p.mkpath}
      end

      # directory with playbook in container
      def container_playbook_path
        dimg.container_dapp_path("ansible-playbook")
      end

    end # Builder::Ansible
  end # Dimg
end # Dapp
