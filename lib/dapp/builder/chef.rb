module Dapp
  module Builder
    # Chef
    class Chef < Base
      LOCAL_COOKBOOK_CHECKSUM_PATTERNS = %w(
        recipes/**/*
        files/**/*
        templates/**/*
      ).freeze

      DEFAULT_CHEFDK_IMAGE = 'dappdeps/chefdk:0.17.3-1'.freeze # TODO: config, DSL, DEFAULT_CHEFDK_IMAGE

      [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
        define_method(:"#{stage}_checksum") { stage_cookbooks_checksum(stage) }

        define_method(:"#{stage}") do |image|
          unless stage_empty?(stage)
            image.add_volumes_from(chefdk_container)
            image.add_commands 'export PATH=/.dapp/deps/chefdk/bin:$PATH'

            image.add_volume "#{stage_tmp_path(stage)}:#{container_stage_tmp_path(stage)}"
            image.add_commands ['chef-solo',
                                "-c #{container_stage_config_path(stage)}",
                                "-o #{stage_cookbooks_runlist(stage).join(',')}"].join(' ')
          end
        end
      end

      def chef_cookbooks_checksum
        stage_cookbooks_checksum(:chef_cookbooks)
      end

      def chef_cookbooks(image)
        image.add_volume "#{cookbooks_vendor_path}:#{application.container_dapp_path('chef_vendored_cookbooks')}"
        image.add_commands(
          'mkdir -p /usr/share/dapp/chef_repo',
          ["cp -a #{application.container_dapp_path('chef_vendored_cookbooks')} ",
           '/usr/share/dapp/chef_repo/cookbooks'].join
        )
      end

      private

      def project_name
        cookbook_metadata.name
      end

      def berksfile_path
        application.home_path('Berksfile')
      end

      def berksfile_lock_path
        application.home_path('Berksfile.lock')
      end

      def berksfile
        @berksfile ||= Berksfile.new(application.home_path, berksfile_path)
      end

      def cookbook_metadata_path
        application.home_path('metadata.rb')
      end

      def cookbook_metadata
        @cookbook_metadata ||= CookbookMetadata.new(cookbook_metadata_path)
      end

      def berksfile_lock_checksum
        application.hashsum berksfile_lock_path.read if berksfile_lock_path.exist?
      end

      def local_cookbook_paths_for_checksum
        @local_cookbook_paths_for_checksum ||= berksfile
                                               .local_cookbooks
                                               .values
                                               .map { |cookbook| cookbook[:path] }
                                               .product(LOCAL_COOKBOOK_CHECKSUM_PATTERNS)
                                               .map { |cb, dir| Dir[cb.join(dir)] }
                                               .flatten
                                               .map(&Pathname.method(:new))
      end

      def stage_cookbooks_paths_for_checksum(stage)
        install_stage_cookbooks(stage)
        Dir[stage_cookbooks_path(stage, '**/*')].map(&Pathname.method(:new))
      end

      def stage_cookbooks_checksum_path(stage)
        application.metadata_path("#{cookbooks_checksum}.#{stage}.checksum")
      end

      def stage_cookbooks_checksum(stage)
        if stage_cookbooks_checksum_path(stage).exist?
          stage_cookbooks_checksum_path(stage).read.strip
        else
          checksum = if stage == :chef_cookbooks
                       cookbooks_checksum
                     else
                       application.hashsum [
                         _paths_checksum(stage_cookbooks_paths_for_checksum(stage)),
                         *application.config._chef._modules,
                         stage == :infra_install ? chefdk_image : nil
                       ].compact
                     end

          stage_cookbooks_checksum_path(stage).write "#{checksum}\n"
          checksum
        end
      end

      def cookbooks_checksum
        @cookbooks_checksum ||= application.hashsum [
          berksfile_lock_checksum,
          _paths_checksum(local_cookbook_paths_for_checksum),
          *application.config._chef._modules
        ]
      end

      def chefdk_image
        DEFAULT_CHEFDK_IMAGE # TODO: config, DSL, DEFAULT_CHEFDK_IMAGE
      end

      def chefdk_container_name # FIXME: hashsum(image) or dockersafe()
        chefdk_image.tr('/', '_').tr(':', '_')
      end

      def chefdk_container
        @chefdk_container ||= begin
          if application.shellout("docker inspect #{chefdk_container_name}").exitstatus.nonzero?
            application.log_secondary_process(application.t(code: 'process.chefdk_loading'), short: true) do
              application.shellout(
                ['docker run',
                 '--restart=no',
                 "--name #{chefdk_container_name}",
                 "--volume /.dapp/deps/chefdk #{chefdk_image}",
                 '2>/dev/null'].join(' ')
              )
            end
          end
          chefdk_container_name
        end
      end

      # rubocop:disable Metrics/MethodLength, Metrics/AbcSize
      def install_cookbooks
        volumes_from = chefdk_container
        application.log_secondary_process(application.t(code: 'process.berks_vendor')) do
          ssh_auth_socket_path = nil
          ssh_auth_socket_path = Pathname.new(ENV['SSH_AUTH_SOCK']).expand_path if ENV['SSH_AUTH_SOCK'] && File.exist?(ENV['SSH_AUTH_SOCK'])

          vendor_commands = [
            'mkdir -p ~/.ssh',
            'echo "Host *" >> ~/.ssh/config',
            'echo "    StrictHostKeyChecking no" >> ~/.ssh/config',
            'if [ ! -f Berksfile.lock ] ; then echo "Berksfile.lock not found" 1>&2 ; exit 1 ; fi',
            'cp -a Berksfile.lock /tmp/Berksfile.lock.orig',
            '/.dapp/deps/chefdk/bin/berks vendor /tmp/vendored_cookbooks',
            'export LOCKDIFF=$(diff -u0 Berksfile.lock /tmp/Berksfile.lock.orig)',
            ['if [ "$LOCKDIFF" != "" ] ; then ',
             'cp -a /tmp/Berksfile.lock.orig Berksfile.lock ; ',
             'echo -e "Bad Berksfile.lock\n$LOCKDIFF" 1>&2 ; exit 1 ; fi'].join,
            ["find /tmp/vendored_cookbooks -type f -exec bash -ec '",
             "install -D -o #{Process.uid} -g #{Process.gid} --mode $(stat -c %a {}) {} ",
             "#{_cookbooks_vendor_path}/$(echo {} | sed -e \"s/\\/tmp\\/vendored_cookbooks\\///g\")' \\;"].join
          ]

          application.shellout!(
            ['docker run --rm',
             ("--volume #{ssh_auth_socket_path}:#{ssh_auth_socket_path}" if ssh_auth_socket_path),
             "--volume #{_cookbooks_vendor_path.tap(&:mkpath)}:#{_cookbooks_vendor_path}",
             *berksfile.local_cookbooks
                       .values
                       .map { |cookbook| "--volume #{cookbook[:path]}:#{cookbook[:path]}" },
             "--volumes-from #{volumes_from}",
             "--workdir #{berksfile_path.parent}",
             ("--env SSH_AUTH_SOCK=#{ssh_auth_socket_path}" if ssh_auth_socket_path),
             "dappdeps/berksdeps:0.1.0 bash #{application.shellout_pack(vendor_commands.join(' && '))}"].compact.join(' '),
            log_verbose: application.log_verbose?
          )

          true
        end
      end
      # rubocop:enable Metrics/MethodLength, Metrics/AbcSize

      def _cookbooks_vendor_path
        application.tmp_path(application.config._name, "cookbooks.#{cookbooks_checksum}")
      end

      def cookbooks_vendor_path(*path)
        _cookbooks_vendor_path.tap do |cookbooks_path|
          install_cookbooks unless cookbooks_path.exist?
        end.join(*path)
      end

      # rubocop:disable Metrics/MethodLength, Metrics/AbcSize
      def install_stage_cookbooks(stage)
        @install_stage_cookbooks ||= {}
        @install_stage_cookbooks[stage] ||= true.tap do
          common_paths = proc do |cookbook_path|
            [['metadata.json', 'metadata.json'],
             ["files/#{stage}", 'files/default'],
             ["templates/#{stage}", 'templates/default']].select { |from, _| cookbook_path.join(from).exist? }
          end

          install_paths = Dir[cookbooks_vendor_path('*')]
                          .map(&Pathname.method(:new))
                          .map do |cookbook_path|
            cookbook_name = File.basename cookbook_path
            is_project = (cookbook_name == project_name)
            is_mdapp = cookbook_name.start_with? 'mdapp-'
            mdapp_enabled = is_mdapp && application.config._chef._modules.include?(cookbook_name)

            paths = if is_project
                      recipe_paths = application.config._chef._recipes
                                                .map { |recipe| ["recipes/#{stage}/#{recipe}.rb", "recipes/#{recipe}.rb"] }
                                                .select { |from, _| cookbook_path.join(from).exist? }

                      (recipe_paths + common_paths[cookbook_path]) if recipe_paths.any?
                    elsif is_mdapp && mdapp_enabled
                      recipe_path = "recipes/#{stage}.rb"
                      if cookbook_path.join(recipe_path).exist?
                        [[recipe_path, recipe_path]] + common_paths[cookbook_path]
                      end
                    else
                      [['.', '.']]
                    end

            [cookbook_path, paths] if paths && paths.any?
          end
                          .compact

          stage_cookbooks_path(stage).mkpath
          install_paths.each do |cookbook_path, paths|
            paths.each do |from, to|
              from_path = cookbook_path.join(from)
              to_path = stage_cookbooks_path(stage, cookbook_path.basename, to)
              to_path.parent.mkpath
              FileUtils.cp_r from_path, to_path
            end
          end
        end
      end
      # rubocop:enable Metrics/MethodLength, Metrics/AbcSize

      # rubocop:disable Metrics/AbcSize
      def stage_cookbooks_runlist(stage)
        install_stage_cookbooks(stage)

        @stage_cookbooks_runlist ||= {}
        @stage_cookbooks_runlist[stage] ||= [].tap do |res|
          to_runlist_entrypoint = proc do |cookbook, entrypoint|
            entrypoint_file = stage_cookbooks_path(stage, cookbook, 'recipes', "#{entrypoint}.rb")
            next unless entrypoint_file.exist?
            "#{cookbook}::#{entrypoint}"
          end

          res.concat(application.config._chef._recipes.map do |recipe|
            to_runlist_entrypoint[project_name, recipe]
          end.flatten.compact)

          res.concat(application.config._chef._modules.map do |mod|
            to_runlist_entrypoint[mod, stage]
          end.flatten.compact)
        end
      end
      # rubocop:enable Metrics/AbcSize

      def stage_empty?(stage)
        stage_cookbooks_runlist(stage).empty?
      end

      def stage_cookbooks_path(stage, *path)
        stage_tmp_path(stage, 'cookbooks', *path)
      end

      def install_chef_solo_stage_config(stage)
        @install_chef_solo_stage_config ||= {}
        @install_chef_solo_stage_config[stage] ||= true.tap do
          stage_tmp_path(stage, 'config.rb').write [
            "file_cache_path \"/var/cache/dapp/chef\"\n",
            "cookbook_path \"#{container_stage_tmp_path(stage, 'cookbooks')}\"\n"
          ].join
        end
      end

      def container_stage_config_path(stage, *path)
        install_chef_solo_stage_config(stage)
        container_stage_tmp_path(stage, 'config.rb', *path)
      end

      def stage_tmp_path(stage, *path)
        application.tmp_path(application.config._name, stage).join(*path)
      end

      def container_stage_tmp_path(_stage, *path)
        path.compact.map(&:to_s).inject(Pathname.new('/chef_build'), &:+)
      end

      def _paths_checksum(paths)
        application.hashsum [
          *paths.map(&:to_s).sort,
          *paths.reject(&:directory?)
                .sort
                .reduce(nil) { |a, e| application.hashsum [a, e.read].compact }
        ]
      end
    end
  end
end
