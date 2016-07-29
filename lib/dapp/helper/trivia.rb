module Dapp
  module Helper
    # Trivia
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

      def self.class_to_lowercase(class_name = self.class)
        class_name.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join('_').downcase.to_s
      end
    end # Trivia
  end # Helper
end # Dapp
