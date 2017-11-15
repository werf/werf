module Dapp
  module Dimg
    module Image
      class Scratch < Stage
        include Helper::Tar

        def initialize(**_kwargs)
          super
          @from_archives = []
        end

        def add_archive(*archives)
          @from_archives.concat(archives.flatten)
        end

        def build!(**_kwargs)
          @built_id = dapp.shellout!("docker import #{prepared_change} #{archive}").stdout.strip
        ensure
          FileUtils.rm_rf(tmp_path)
        end

        protected

        attr_accessor :from_archives

        def archive
          tmp_path('archive.tar').tap do |archive_path|
            tar_write(archive_path) do |common_tar|
              from_archives.each do |from_archive|
                tar_gz_read(from_archive) do |tar|
                  tar.each_entry do |entry|
                    mode = entry.header.mode
                    path = entry.full_name

                    if entry.directory?
                      common_tar.mkdir path, mode
                    elsif entry.symlink?
                      common_tar.add_symlink path, entry.header.linkname, mode
                    else
                      common_tar.add_file path, mode do |tf|
                        tf.write entry.read
                      end
                    end
                  end
                end
              end
            end
          end
        end

        def tmp_path(*path)
          @tmp_path ||= Dir.mktmpdir('dapp-scratch-', dapp.tmp_base_dir)
          dapp.make_path(@tmp_path, *path).expand_path.tap { |p| p.parent.mkpath }
        end
      end # Stage
    end # Image
  end # Dimg
end # Dapp
