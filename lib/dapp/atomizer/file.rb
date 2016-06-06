module Dapp
  module Atomizer
    class File < Base
      def rollback(files)
        FileUtils.rm_rf files
      end
    end
  end
end
