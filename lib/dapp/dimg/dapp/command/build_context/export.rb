module Dapp
  module Dimg
    module Dapp
      module Command
        module BuildContext
          module Export
            def build_context_export
              log_process(:'export context') do
                export_build_context_image_tar
                export_build_context_build_tar
              end
            end

            def export_build_context_image_tar
              lock("#{name}.images", readonly: true) do
                context_images_names = build_configs.map do |config|
                  Dimg.new(config: config, dapp: self).tagged_images.map(&:name)
                end.flatten

                log_secondary_process(:images, short: true) do
                  Image::Docker.save!(context_images_names, build_context_images_tar, verbose: true, quiet: log_quiet?)
                end unless context_images_names.empty?
              end
            end

            def export_build_context_build_tar
              log_secondary_process(:build_dir, short: true) do
                tar_write(build_context_build_tar) do |tar|
                  Dir.glob(File.join(build_path, '**/*'), File::FNM_DOTMATCH).each do |path|
                    archive_file_path = path
                                          .reverse
                                          .chomp(build_path.to_s.reverse)
                                          .chomp('/')
                                          .reverse
                    if File.directory?(path)
                      tar.mkdir archive_file_path, File.stat(path).mode
                    else
                      tar.add_file archive_file_path, File.stat(path).mode do |tf|
                        tf.write File.read(path)
                      end
                    end
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
