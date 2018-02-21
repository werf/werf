module Dapp
  module Helper
    module Trivia
      def kwargs(args)
        args.last.is_a?(Hash) ? args.pop : {}
      end

      def class_to_lowercase(class_name = self.class)
        Trivia.class_to_lowercase(class_name)
      end

      def delete_file(path)
        path = Pathname(path)
        path.delete if path.exist?
      end

      def search_file_upward(filename)
        cdir = Pathname(work_dir)
        loop do
          if (path = cdir.join(filename)).exist?
            return path.to_s
          end
          break if (cdir = cdir.parent).root?
        end
      end

      def make_path(base, *path)
        Pathname.new(File.join(base.to_s, *path.compact.map(&:to_s)))
      end

      def ignore_path?(path, paths: [], exclude_paths: [])
        ignore_path_base(path, exclude_paths: exclude_paths) do
          paths.empty? ||
            paths.any? do |p|
              File.fnmatch?(p, path, File::FNM_PATHNAME|File::FNM_DOTMATCH) ||
                File.fnmatch?(File.join(p, '**', '*'), path, File::FNM_PATHNAME|File::FNM_DOTMATCH)
            end
        end
      end

      def ignore_path_base(path, exclude_paths: [])
        is_exclude_path = exclude_paths.any? { |p| check_path?(path, p) }
        is_include_path = yield
        is_exclude_path || !is_include_path
      end

      def check_path?(path, format)
        path_checker(path) { |checking_path| File.fnmatch(format, checking_path, File::FNM_PATHNAME|File::FNM_DOTMATCH) }
      end

      def check_subpath?(path, format)
        path_checker(format) { |checking_path| File.fnmatch(checking_path, path, File::FNM_PATHNAME|File::FNM_DOTMATCH) }
      end

      def path_checker(path)
        path_parts = path.split('/')
        checking_path = nil

        until path_parts.empty?
          checking_path = [checking_path, path_parts.shift].compact.join('/')
          return true if yield checking_path
        end
        false
      end

      def self.class_to_lowercase(class_name = self)
        class_name.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join('_').downcase.to_s
      end
    end # Trivia
  end # Helper
end # Dapp
