module Dapp
  module Dimg
    module Dapp
      module Command
        module BuildContext
          module Import
            def build_context_import
              raise ::Dapp::Error::Command, code: :context_directory_not_found,
                                            data: { path: build_context_path } unless build_context_path.exist?

              log_process(:'import context') do
                import_build_context_build_tar
                import_build_context_image_tar
              end
            end

            def import_build_context_image_tar
              if build_context_images_tar.exist?
                log_secondary_process(:images) do
                  lock("#{name}.images") do
                    Image::Stage.load!(self, build_context_images_tar, verbose: true, quiet: log_quiet?)
                  end unless dry_run?
                end
              else
                log_warning(desc: { code: :context_archive_not_found, data: { path: build_context_images_tar } })
              end
            end

            def import_build_context_build_tar
              if build_context_build_tar.exist?
                log_secondary_process(:build_dir, short: true) do
                  unless dry_run?
                    store_current_build_dir

                    if !!options[:use_system_tar]
                      FileUtils.mkpath build_path
                      shellout!("tar -xf #{build_context_build_tar} -C #{build_path}")
                    else
                      tar_read(build_context_build_tar) do |tar|
                        tar.each_entry do |entry|
                          header = entry.header
                          path = File.join(build_path, entry.full_name)

                          if entry.directory?
                            FileUtils.mkpath path, :mode => entry.header.mode
                          else
                            FileUtils.mkpath File.dirname(path)
                            File.write(path, entry.read)
                            File.chmod(header.mode, path)
                          end
                        end
                      end
                    end
                  end
                end
              else
                log_warning(desc: { code: :context_archive_not_found, data: { path: build_context_build_tar } })
              end
            end

            def store_current_build_dir
              return if build_path_empty?
              raise ::Dapp::Error::Command, code: :stored_build_dir_already_exist,
                                            data: { path: "#{build_path}.old" } if File.exist?("#{build_path}.old")
              FileUtils.mv(build_path, "#{build_path}.old")
            end

            def build_path_empty?
              (build_path.entries.map(&:to_s) - %w(. .. locks)).empty?
            end
          end
        end
      end
    end
  end
end
