module Dapp
  module Dimg
    module Dapp
      module Command
        module BuildContext
          module Import
            def build_context_import
              raise Error::Command, code: :context_directory_not_found,
                                    data: { path: build_context_path } unless build_context_path.exist?

              log_process(:'import context') do
                with_log_indent do
                  import_build_context_build_tar
		              import_build_context_image_tar
                end
              end
            end

            def import_build_context_image_tar
              if build_context_images_tar.exist?
                log_secondary_process(:images, short: true) do
                  lock("#{name}.images") do
                    Image::Docker.load!(build_context_images_tar)
                  end
                end
              else
                log_warning(desc: { code: :context_archive_not_found, data: { path: build_context_images_tar } })
              end
            end

            def import_build_context_build_tar
              if build_context_build_tar.exist?
                log_secondary_process(:build_dir, short: true) do
                  store_current_build_dir

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
              else
                log_warning(desc: { code: :context_archive_not_found, data: { path: build_context_build_tar } })
              end
            end

            def store_current_build_dir
              return if build_path_empty?
              raise Error::Command, code: :stored_build_dir_already_exist,
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
