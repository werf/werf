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

      def self.class_to_lowercase(class_name = self.class)
        class_name.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join('_').downcase.to_s
      end
    end # Trivia
  end # Helper
end # Dapp
