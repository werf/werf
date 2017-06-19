module Dapp
  module Helper
    module Tar
      def tar_write(path)
        File.open(path, File::RDWR | File::CREAT) do |f|
          Gem::Package::TarWriter.new(f) do |tar|
            yield tar if block_given?
          end
        end
      end

      def tar_read(path)
        File.open(path, File::RDONLY) do |f|
          Gem::Package::TarReader.new(f) do |tar|
            yield tar if block_given?
          end
        end
      end

      def tar_gz_read(path)
        File.open(path, File::RDONLY) do |f_gz|
          Zlib::GzipReader.wrap(f_gz) do |f|
            Gem::Package::TarReader.new(f) do |tar|
              yield tar if block_given?
            end
          end
        end
      end
    end # Tar
  end # Helper
end # Dapp
