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
        recipes/%{stage}.rb
        recipes/*_%{stage}.rb
        files/%{stage}/*
        templates/%{stage}/*
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

      private

      def project_name
        application.config._root_app._name # FIXME: parse name from metadata.rb
      end

      def berksfile_path
        application.home_path('Berksfile')
      end

      def berksfile
        @berksfile ||= Berksfile.new(application.home_path, berksfile_path)
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

          res.concat(application.config._app_runlist.map(&:_name).map do |name|
            subname_parts = name.split(project_name, 2)[1].split('-')
            to_runlist_entrypoint[project_name, [*subname_parts, stage].join('_')]
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
          .map do|cookbook_path|
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

          application.hashsum([_paths_checksum(stage_cookbooks_vendored_paths(stage, with_files: true)),
                               *application.config._chef._modules,
                               (stage == :infra_install) ? chefdk_image : nil].compact).tap do |checksum|
            stage_cookbooks_checksum_path(stage).write "#{checksum}\n"
          end
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
          if application.shellout("docker inspect #{chefdk_container_name}").exitstatus != 0
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

      # rubocop:disable Metrics/MethodLength
      def install_cookbooks
        @install_cookbooks ||= begin
          volumes_from = chefdk_container
          application.log_secondary_process(application.t(code: 'process.berks_vendor')) do
            application.shellout!(
              ['docker run --rm',
               "--volumes-from #{volumes_from}",
               "--volume #{cookbooks_vendor_path.tap(&:mkpath)}:#{cookbooks_vendor_path}",
               *berksfile.local_cookbooks
                         .values
                         .map { |cookbook| "--volume #{cookbook[:path]}:#{cookbook[:path]}" },
               "--user #{Process.uid}:#{Process.gid}",
               "--workdir #{berksfile_path.parent}",
               '--env BERKSHELF_PATH=/tmp/berkshelf',
               "ubuntu:14.04 /.dapp/deps/chefdk/bin/berks vendor #{cookbooks_vendor_path}"
              ].join(' '),
              log_verbose: application.log_verbose?
            )

            true
          end
        end
      end
      # rubocop:enable Metrics/MethodLength

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
