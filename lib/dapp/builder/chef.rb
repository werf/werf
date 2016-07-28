module Dapp
  module Builder
    # Chef
    class Chef < Base
      LOCAL_COOKBOOK_PATTERNS = %w(
        recipes/**/*
        files/**/*
        templates/**/*
      ).freeze

      STAGE_LOCAL_COOKBOOK_PATTERNS = %w(
        metadata.json
        recipes/%{stage}.rb
        recipes/*_%{stage}.rb
        files/default/%{stage}/*
        templates/default/%{stage}/*
      ).freeze

      DEFAULT_CHEFDK_IMAGE = 'dappdeps/chefdk:0.15.16-3'.freeze # TODO: config, DSL, DEFAULT_CHEFDK_IMAGE

      [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
        define_method(:"#{stage}_checksum") { stage_cookbooks_checksum(stage) }

        define_method(:"#{stage}") do |image|
          install_stage_cookbooks(stage)
          install_chef_solo_stage_config(stage)

          unless stage_empty?(stage)
            image.add_volumes_from(chefdk_container)
            image.add_commands 'export PATH=/.dapp/deps/chefdk/bin:$PATH'

            image.add_volume "#{stage_build_path(stage)}:#{container_stage_build_path(stage)}"
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
        image.add_commands(
          "mkdir -p /usr/share/dapp/chef_repo",
          "cp -a #{container_cookbooks_vendor_path} /usr/share/dapp/chef_repo/cookbooks"
        )
      end

      private

      def project_name
        cookbook_metadata.name
      end

      def berksfile_path
        application.home_path('Berksfile')
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
        path = application.home_path('Berksfile.lock')
        application.hashsum path.read if path.exist?
      end

      # rubocop:disable Metrics/AbcSize
      def stage_cookbooks_runlist(stage)
        @stage_cookbooks_runlist ||= {}
        @stage_cookbooks_runlist[stage] ||= [].tap do |res|
          to_runlist_entrypoint = proc do |name, entrypoint|
            entrypoint_file = stage_cookbooks_path(stage, name, 'recipes', "#{entrypoint}.rb")
            next unless entrypoint_file.exist?
            "#{name}::#{entrypoint}"
          end

          res.concat(application.config._chef._modules.map do |name|
            to_runlist_entrypoint[name, stage]
          end.compact)

          project_main_entry = to_runlist_entrypoint[project_name, stage]
          res << project_main_entry if project_main_entry

          res.concat(application.config._app_runlist.map do |app_component|
            to_runlist_entrypoint[project_name, [app_component, stage].join('_')]
          end.compact)
        end
      end
      # rubocop:enable Metrics/AbcSize

      def local_cookbook_paths
        @local_cookbook_paths ||= berksfile.local_cookbooks
                                           .values
                                           .map { |cookbook| cookbook[:path] }
                                           .product(LOCAL_COOKBOOK_PATTERNS)
                                           .map { |cb, dir| Dir[cb.join(dir)] }
                                           .flatten
                                           .map(&Pathname.method(:new))
      end

      def stage_cookbooks_vendored_paths(stage, with_files: false)
        Dir[cookbooks_vendor_path('*')]
          .map do |cookbook_path|
            if ['mdapp-*', project_name].any? { |pattern| File.fnmatch(pattern, File.basename(cookbook_path)) }
              STAGE_LOCAL_COOKBOOK_PATTERNS.map do |pattern|
                Dir[File.join(cookbook_path, pattern % { stage: stage })]
              end
            elsif with_files
              Dir[File.join(cookbook_path, '**/*')]
            else
              cookbook_path
            end
          end
          .flatten
          .map(&Pathname.method(:new))
      end

      def stage_cookbooks_checksum_path(stage)
        application.build_cache_path("#{cookbooks_checksum}.#{stage}.checksum")
      end

      def stage_cookbooks_checksum(stage)
        if stage_cookbooks_checksum_path(stage).exist?
          stage_cookbooks_checksum_path(stage).read.strip
        else
          install_cookbooks

          if stage == :chef_cookbooks
            checksum = cookbooks_checksum
          else
            checksum = [_paths_checksum(stage_cookbooks_vendored_paths(stage, with_files: true)),
                        *application.config._chef._modules,
                        stage == :infra_install ? chefdk_image : nil].compact
          end

          stage_cookbooks_checksum_path(stage).write "#{checksum}\n"
          checksum
        end
      end

      def cookbooks_checksum
        @cookbooks_checksum ||= application.hashsum [
          berksfile_lock_checksum,
          _paths_checksum(local_cookbook_paths),
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
        @install_cookbooks ||= begin
          volumes_from = chefdk_container
          application.log_secondary_process(application.t(code: 'process.berks_vendor')) do
            ssh_auth_socket_path = nil
            ssh_auth_socket_path = Pathname.new(ENV['SSH_AUTH_SOCK']).expand_path if ENV['SSH_AUTH_SOCK'] && File.exist?(ENV['SSH_AUTH_SOCK'])

            application.shellout!(
              ['docker run --rm',
               '--volume /etc:/etc:ro',
               '--volume /usr:/usr:ro',
               '--volume /lib:/lib:ro',
               '--volume /lib64:/lib64:ro',
               '--volume /home:/home',
               '--volume /tmp:/tmp',
               ("--volume #{ssh_auth_socket_path.dirname}:#{ssh_auth_socket_path.dirname}" if ssh_auth_socket_path),
               "--volume #{cookbooks_vendor_path.tap(&:mkpath)}:#{cookbooks_vendor_path}",
               *berksfile.local_cookbooks
                         .values
                         .map { |cookbook| "--volume #{cookbook[:path]}:#{cookbook[:path]}" },
               "--volumes-from #{volumes_from}",
               "--user #{Process.uid}:#{Process.gid}",
               "--workdir #{berksfile_path.parent}",
               '--env BERKSHELF_PATH=/tmp/berkshelf',
               ("--env SSH_AUTH_SOCK=#{ssh_auth_socket_path}" if ssh_auth_socket_path),
               "dappdeps/berksdeps:0.1.0 /.dapp/deps/chefdk/bin/berks vendor #{cookbooks_vendor_path}"].compact.join(' '),
              log_verbose: application.log_verbose?
            )

            true
          end
        end
      end
      # rubocop:enable Metrics/MethodLength, Metrics/AbcSize

      def install_stage_cookbooks(stage)
        stage_cookbooks_path(stage).mkpath
        stage_cookbooks_vendored_paths(stage).each do |path|
          new_path = stage_cookbooks_path(stage, path.relative_path_from(cookbooks_vendor_path))
          new_path.parent.mkpath
          FileUtils.cp_r path, new_path
        end
      end

      def stage_empty?(stage)
        stage_cookbooks_runlist(stage).empty?
      end

      def install_chef_solo_stage_config(stage)
        stage_config_path(stage).write [
          "file_cache_path \"/var/cache/dapp/chef\"\n",
          "cookbook_path \"#{container_stage_cookbooks_path(stage)}\"\n"
        ].join
      end

      def cookbooks_vendor_path(*path)
        application.build_path('chef', 'vendored_cookbooks').join(*path)
      end

      def container_cookbooks_vendor_path(*path)
        application.container_build_path('chef', 'vendored_cookbooks').join(*path)
      end

      def stage_build_path(stage, *path)
        application.build_path(application.config._name, stage).join(*path)
      end

      def container_stage_build_path(_stage, *path)
        path.compact.map(&:to_s).inject(Pathname.new('/chef_build'), &:+)
      end

      def stage_cookbooks_path(stage, *path)
        stage_build_path(stage, 'cookbooks', *path)
      end

      def container_stage_cookbooks_path(stage, *path)
        container_stage_build_path(stage, 'cookbooks', *path)
      end

      def stage_config_path(stage, *path)
        stage_build_path(stage, 'config.rb', *path)
      end

      def container_stage_config_path(stage, *path)
        container_stage_build_path(stage, 'config.rb', *path)
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
