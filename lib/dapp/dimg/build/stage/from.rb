module Dapp
  module Dimg
    module Build
      module Stage
        class From < Base
          def dependencies
            @dependencies ||= [from_image_name, dimg.config._docker._from_cache_version, config_mounts_dirs]
          end

          protected

          def prepare_image
            from_image.pull!
            raise Error::Build, code: :from_image_not_found, data: { name: from_image_name } unless from_image.tagged?
            add_cleanup_mounts_dirs_command
            super
          end

          def add_cleanup_mounts_dirs_command
            return if config_mounts_dirs.empty?
            image.add_service_command ["#{dimg.dapp.rm_bin} -rf %s",
                                       "#{dimg.dapp.mkdir_bin} -p %s"].map { |c| format(c, config_mounts_dirs.join(' ')) }
          end

          def config_mounts_dirs
            ([:tmp_dir, :build_dir].map { |type| config_mounts_by_type(type) } + config_custom_dir_mounts.map(&:last)).flatten.uniq
          end

          def adding_mounts_by_type(_type)
            []
          end

          def adding_custom_dir_mounts
            []
          end

          def image_should_be_untagged_condition
            false
          end

          def should_not_be_detailed?
            from_image.tagged?
          end

          private

          def from_image_name
            dimg.config._docker._from
          end

          def from_image
            @from_image ||= Image::Stage.image_by_name(name: from_image_name, dapp: dimg.dapp)
          end
        end # Prepare
      end # Stage
    end # Build
  end # Dimg
end # Dapp
