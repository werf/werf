module Dapp
  module Helper
    # Sha256
    module Sha256
      def hashsum(arg)
        sha256(arg)
      end

      def paths_content_hashsum(paths)
        paths.map(&:to_s)
             .reject { |path| File.directory?(path) }
             .sort
             .reduce(nil) { |hash, path| hashsum [hash, File.read(path)].compact }
      end

      def sha256(arg)
        Digest::SHA256.hexdigest Array(arg).compact.map(&:to_s).join(':::')
      end
    end
  end # Helper
end # Dapp
