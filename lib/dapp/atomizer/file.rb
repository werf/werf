module Dapp
  module Atomizer
    class File < Base
      def del_path(path)
        path, cached_path = cp_paths(path)
        if path.exist?
          send(:<<, "del #{path}")
          FileUtils.cp path, cached_path, preserve: true
        end
      end

      def add_path(path)
        send(:<<, path)
      end

      def rollback(lines)
        deleted_files, added_files = lines.partition { |line| line.start_with?('del') }
        FileUtils.rm_rf added_files
        deleted_files.each do |line|
          path, cached_path = cp_paths(line.split.last)
          FileUtils.mv cached_path, path if cached_path.exist?
        end
      end

      protected

      def cp_paths(path)
        path = Pathname(path)
        [path, Pathname(::File.join(file_path.dirname, "#{file_path.basename}.#{path.basename}"))]
      end
    end
  end
end
