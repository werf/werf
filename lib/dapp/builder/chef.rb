module Dapp
  module Builder
    class Chef < Base
      COOKBOOK_PATTERNS = %w(
        recipes/**/*
        files/**/*
        templates/**/*
      )

      COOKBOOK_STAGE_PATTERNS = %w(
        recipes/%{stage}-*.rb
        files/%{stage}/*
        templates/%{stage}/*
      )

      [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
        define_method(:"#{stage}_checksum") {berks_vendor_stage_checksum(stage)}

        define_method(:"#{stage}") do |image|
          install_berks_vendor_stage(stage)
          install_chef_solo_stage_config(stage)

          image.add_volume '/opt/chefdk:/opt/chefdk'
          image.add_volume "#{stage_cookbooks_path(stage)}:#{container_stage_cookbooks_path(stage)}"
          image.add_commands(
            "mkdir -p #{container_chef_path}",
            "/opt/chefdk/sbin/chef-solo -c #{container_chef_solo_stage_config_path(stage)}",
          )
        end
      end

      private

      def berksfile
        @berksfile ||= Dapp::Berksfile.new(application, application.home_path('Berksfile'))
      end

      def berksfile_lock_path
        application.home_path('Berksfile.lock')
      end

      def berksfile_lock_checksum
        hashsum berksfile_lock_path.read if berksfile_lock_path.exist?
      end

      def local_cookbook_paths
        @local_cookbook_paths ||= berksfile.local_cookbook_paths
          .product(COOKBOOK_PATTERNS)
          .map {|cb, dir| Dir[cb.join(dir)]}
          .flatten
          .map(&Pathname.method(:new))
          .sort
      end

      def berks_vendor_checksum
        @berks_vendor_checksum ||= hashsum(berksfile_lock_checksum,
                                           *local_cookbook_paths.map(&:to_s),
                                           *local_cookbook_paths.reject(&:directory?).map(&:read))
      end

      def berks_vendor_stage_paths(stage)
        @berks_vendor_stage_paths ||= {}
        @berks_vendor_stage_paths[stage] ||= COOKBOOK_STAGE_PATTERNS
          .map {|pattern| Dir[berks_vendor_path('*', pattern % {stage: stage})]}
          .flatten
          .map(&Pathname.method(:new))
          .sort
      end

      def berks_vendor_stage_checksum_path(stage)
        application.build_cache_path("#{berks_vendor_checksum}.#{stage}.checksum")
      end

      def berks_vendor_stage_checksum(stage)
        if berks_vendor_stage_checksum_path(stage).exist?
          berks_vendor_stage_checksum_path(stage).read.strip
        else
          install_berks_vendor

          hashsum(*berks_vendor_stage_paths(stage).map(&:to_s),
                  *berks_vendor_stage_paths(stage).reject(&:directory?).map(&:read)
                 ).tap do |checksum|
            berks_vendor_stage_checksum_path(stage).write "#{checksum}\n"
          end
        end
      end

      def install_berks_vendor
        shellout! "berks vendor #{berks_vendor_path.tap(&:mkdir_p)}"
      end

      def install_berks_vendor_stage(stage)
        berks_vendor_stage_checksum(stage)

        berks_vendor_stage_paths(stage).each do |path|
          new_path = stage_cookbooks_path(stage).join(path.relative_path_from(berks_vendor_path))
          new_path.parent.mkdir_p
          FileUtils.cp path, new_path
        end
      end

      def chef_path(*path)
        application.build_path('chef', *path)
      end

      def container_chef_path(*path)
        Pathname.new('/usr/share/dapp/chef')
      end

      def berks_vendor_path(*path)
        chef_path('vendored_cookbooks', *path)
      end

      def stage_cookbooks_path(stage, *path)
        chef_path("#{stage}_cookbooks", *path)
      end

      def container_stage_cookbooks_path(stage, *path)
        container_chef_path("#{stage}_cookbooks", *path)
      end

      def chef_solo_stage_config_path(stage)
        chef_path("#{stage}_chef_solo.rb")
      end

      def container_chef_solo_stage_config_path(stage)
        container_chef_path("#{stage}_chef_solo.rb")
      end

      def install_chef_solo_stage_config(stage)
        chef_solo_stage_config_path(stage).write [
          "cookbook_path #{container_stage_cookbooks_path(stage)}\n",
        ].join
      end
    end
  end
end

