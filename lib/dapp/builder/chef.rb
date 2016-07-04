module Dapp
  module Builder
    class Chef < Base
      LOCAL_COOKBOOK_PATTERNS = %w(
        recipes/**/*
        files/**/*
        templates/**/*
      )

      STAGE_COOKBOOK_PATTERNS = %w(
        recipes/%{stage}-*.rb
        files/%{stage}/*
        templates/%{stage}/*
      )

      [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
        define_method(:"#{stage}_checksum") {stage_cookbooks_checksum(stage)}

        define_method(:"#{stage}") do |image|
          install_stage_cookbooks(stage)
          install_chef_solo_stage_config(stage)

          unless stage_empty?(stage)
            image.add_volume '/opt/chefdk:/opt/chefdk'
            image.add_volume "/tmp/dapp/chef_cache_#{SecureRandom.uuid}:/var/cache/dapp/chef"
            image.add_volume "#{stage_build_path(stage)}:#{container_stage_build_path(stage)}"

            image.add_commands ["/opt/chefdk/chef-solo",
                                "-c #{container_stage_config_path(stage)}",
                               ].join(' ')
          end
        end
      end

      private

      def berksfile
        @berksfile ||= Dapp::Berksfile.new(application, application.home_path('Berksfile'))
      end

      def berksfile_lock_checksum
        path = application.home_path('Berksfile.lock')
        application.hashsum path.read if path.exist?
      end

      def local_cookbook_paths
        @local_cookbook_paths ||= berksfile.local_cookbook_paths
          .product(LOCAL_COOKBOOK_PATTERNS)
          .map {|cb, dir| Dir[cb.join(dir)]}
          .flatten
          .map(&Pathname.method(:new))
          .sort
      end

      def stage_cookbooks_paths(stage)
        @stage_cookbooks_paths ||= {}
        @stage_cookbooks_paths[stage] ||= STAGE_COOKBOOK_PATTERNS
          .map {|pattern| Dir[cookbooks_path('*', pattern % {stage: stage})]}
          .flatten
          .map(&Pathname.method(:new))
          .sort
      end

      def stage_cookbooks_checksum_path(stage)
        application.build_cache_path("#{cookbooks_checksum}.#{stage}.checksum")
      end

      def stage_cookbooks_checksum(stage)
        if stage_cookbooks_checksum_path(stage).exist?
          stage_cookbooks_checksum_path(stage).read.strip
        else
          install_cookbooks

          application.hashsum([*stage_cookbooks_paths(stage).map(&:to_s),
                               *stage_cookbooks_paths(stage).reject(&:directory?).map(&:read)
                              ]).tap do |checksum|
            stage_cookbooks_checksum_path(stage).write "#{checksum}\n"
          end
        end
      end

      def cookbooks_checksum
        @cookbooks_checksum ||= application.hashsum [
          berksfile_lock_checksum,
          *local_cookbook_paths.map(&:to_s),
          *local_cookbook_paths.reject(&:directory?).map(&:read),
        ]
      end

      def install_cookbooks
        @install_cookbooks ||= begin
          application.shellout!(["cd #{application.home_path}",
                                 "berks vendor #{cookbooks_path.tap(&:mkpath)}"].join(' && '),
                                log_verbose: true)
          true
        end
      end

      def install_stage_cookbooks(stage)
        stage_cookbooks_path(stage).mkpath
        stage_cookbooks_paths(stage).each do |path|
          new_path = stage_cookbooks_path(stage, path.relative_path_from(cookbooks_path))
          new_path.parent.mkpath
          FileUtils.cp path, new_path
        end
      end

      def stage_empty?(stage)
        (not stage_cookbooks_path(stage).exist?) or
          stage_cookbooks_path(stage).entries.size <= 2
      end

      def install_chef_solo_stage_config(stage)
        stage_config_path(stage).write [
          "file_cache_path \"/var/cache/dapp/chef\"",
          "cookbook_path \"#{container_stage_cookbooks_path(stage)}\"",
        ].join("\n")
      end


      def cookbooks_path(*path)
        application.build_path('chef', 'vendored_cookbooks', *path)
      end

      def stage_build_path(stage, *path)
        application.build_path('chef', stage, *path)
      end

      def container_stage_build_path(stage, *path)
        path.compact.inject(Pathname.new('/chef_build'), &:+)
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
    end
  end
end

