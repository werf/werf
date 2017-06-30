module Dapp
  module Error
    class ImageBuildFailed < Build
      def initialize(net_status={})
        super( {context: :build, code: :image_build_failed}.merge(net_status) )
      end
    end
  end
end
