module Dapp
  module Dimg
    class Builder::Ansible < Builder::Base

      ANSIBLE_IMAGE_VERSION = "2.4.4.0-10"

      def ansible_bin
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/embedded/bin/ansible"
      end

      def ansible_playbook_bin
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/embedded/bin/ansible-playbook"
      end

      def python_path
        "/.dapp/deps/ansible/#{ANSIBLE_IMAGE_VERSION}/embedded/bin/python"
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

      # query tasks from ansible config
      # create dump_config structure
      # returns structure:
      # { 'tasks' => [array of tasks for stage],
      #   'dump_config' => {
      #      'dump_config_doc' => 'dump of doc',
      #      'dump_config_sections' => {'task_0'=>'dump for task 0', 'task_1'=>'dump for task 1', ... }}
      def stage_config(stage)
        @stage_configs ||= {}
        @stage_configs[stage.to_s] ||= begin
          {}.tap do |stage_config|
            stage_config['dump_config'] = {
              'dump_config_doc' => dimg.config._ansible['dump_config_doc'],
              'dump_config_sections' => {},
            }
            stage_config['tasks'] = dimg.config._ansible[stage.to_s].map.with_index do |dapp_task, task_num|
              {}.tap do |task|
                task.merge!(dapp_task['config'])
                task['tags'] = [].tap do |tags|
                  tags << dapp_task['config']['tags']
                  dump_tag = "task_#{task_num}"
                  tags << dump_tag
                  stage_config['dump_config']['dump_config_sections'][dump_tag] = dapp_task['dump_config_section']
                end.flatten.compact
              end
            end || []
          end
        end
      end

      def stage_playbook(stage)
        @stage_playbooks ||= {}
        @stage_playbooks[stage.to_s] ||= begin
          [{
            'hosts' => 'all',
            'gather_facts' => 'no',
            'tasks' => [],
           }].tap do |playbook|
            playbook[0]['tasks'].concat(stage_config(stage)['tasks'])
          end
        end
      end

      def create_workdir_structure(stage)
        @workdir_structures ||= {}
        @workdir_structures[stage.to_s] ||= true.tap do
          host_workdir(stage).tap do |workdir|
            # playbook with tasks for a stage
            workdir.join('playbook.yml').write YAML.dump(stage_playbook(stage))
            #puts YAML.dump(stage_playbook(stage))

            # generate inventory with localhost and python in dappdeps-ansible
            workdir.join('hosts').write Assets.hosts(python_path)

            # generate ansible config for solo mode
            workdir.join('ansible.cfg').write Assets.ansible_cfg(container_workdir.join('hosts'),
                                                                 container_workdir.join('lib', 'callback'),
                                                                 dimg.dapp.sudo_bin,
                                                                 container_tmpdir.join('local'),
                                                                 container_tmpdir.join('remote'),
                                                                 )

            # save config dump for pretty errors
            workdir.join('dump_config.json').write JSON.generate(stage_config(stage)['dump_config'])

            # python modules
            workdir.join('lib').tap do |libdir|
              libdir.mkpath
              # crypt.py hack
              # TODO must be in dappdeps-ansible
              libdir.join('crypt.py').write Assets.crypt_py
              libdir.join('callback').tap do |callbackdir|
                callbackdir.mkpath
                callbackdir.join('__init__.py').write '# module callback'
                callbackdir.join('live.py').write Assets.live_py
                # add dapp specific stdout callback for ansible
                callbackdir.join('dapp.py').write Assets.dapp_py
              end
            end
          end
        end
      end

      %i(before_install install before_setup setup build_artifact).each do |stage|
        define_method("#{stage}?") {
          !stage_empty?(stage)
        }

        define_method("#{stage}_checksum") do
          checksum_args = []
          checksum_args << JSON.dump(stage_config(stage)['tasks']) unless stage_config(stage)['tasks'].empty?
          checksum_args << public_send("#{stage}_version_checksum")
          _checksum checksum_args
        end

        define_method("#{stage}_version_checksum") do
          _checksum(dimg.config._ansible["#{stage}_version"], dimg.config._ansible['version'])
        end

        define_method(stage.to_s) do |image|
          unless stage_empty?(stage)
            create_workdir_structure(stage)
            image.add_env('ANSIBLE_CONFIG', container_workdir.join('ansible.cfg'))
            image.add_env('DAPP_DUMP_CONFIG_DOC_PATH', container_workdir.join('dump_config.json'))
            image.add_env('PYTHONPATH', container_workdir.join('lib'))
            image.add_env('PYTHONIOENCODING', 'utf-8')
            image.add_env('ANSIBLE_PREPEND_SYSTEM_PATH', dimg.dapp.dappdeps_base_path)
            image.add_env('LC_ALL', 'C.UTF-8')
            image.add_volumes_from("#{ansible_container}:rw")
            image.add_volume "#{host_workdir(stage)}:#{container_workdir}:ro"
            image.add_volume "#{host_tmpdir(stage)}:#{container_tmpdir}:rw"
            image.add_command [ansible_playbook_bin,
                               container_workdir.join('playbook.yml'),
                               ENV['ANSIBLE_ARGS']
                              ].compact.join(' ')
          end
        end
      end

      def before_build_check
      end

      def before_dimg_should_be_built_check
      end

      def stage_empty?(stage)
        stage_config(stage)['tasks'].empty? && public_send("#{stage}_version_checksum").nil?
      end

      # host directory in tmp_dir with directories structure
      def host_workdir(stage)
        @host_workdirs ||= {}
        @host_workdirs[stage.to_s] ||= begin
          dimg.tmp_path(dimg.dapp.consistent_uniq_slugify(dimg.config._name || "nameless"), "ansible-workdir-#{stage}").tap {|p| p.mkpath}
        end
      end

      # temporary directories for ansible
      def host_tmpdir(stage)
        @host_tmpdirs ||= {}
        @host_tmpdirs[stage.to_s] ||= begin
          dimg.tmp_path(dimg.dapp.consistent_uniq_slugify(dimg.config._name || "nameless"), "ansible-tmpdir-#{stage}").tap do |p|
            p.mkpath
            p.join('local').mkpath
            p.join('remote').mkpath
          end
        end
      end

      # directory with playbook in container
      def container_workdir
        dimg.container_dapp_path("ansible-workdir")
      end

      # temporary directory for ansible
      def container_tmpdir
        dimg.container_dapp_path("ansible-tmpdir")
      end

    end # Builder::Ansible
  end # Dimg
end # Dapp
