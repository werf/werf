module Dapp
  module Atomizer
    class Docker < Base
      def rollback(images)
        Mixlib::ShellOut.new("docker rmi #{images.join(' ')}").run_command
      end
    end
  end
end
