module Dapp
  module CommonHelper
    def sha256(arg)
      Digest::SHA256.hexdigest Array(arg).map(&:to_s).join(':::')
    end
  end # CommonHelper
end # Dapp
