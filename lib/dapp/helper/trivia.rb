module Dapp
  module Helper
    # Trivia
    module Trivia
      def kwargs(args)
        args.last.is_a?(Hash) ? args.pop : {}
      end

      def delete_file(path)
        path = Pathname(path)
        path.delete if path.exist?
      end

      def to_mb(bytes)
        (bytes / 1024.0 / 1024.0).round(2)
      end
    end # Trivia
  end # Helper
end # Dapp
