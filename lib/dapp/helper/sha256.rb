module Dapp
  module Helper
    # Sha256
    module Sha256
      def hashsum(arg)
        sha256(arg)
      end

      def sha256(arg)
        Digest::SHA256.hexdigest Array(arg).compact.map(&:to_s).join(':::')
      end
    end
  end # Helper
end # Dapp
