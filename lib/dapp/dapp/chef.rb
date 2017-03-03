module Dapp
  class Dapp
    module Chef
      def local_git_artifact_exclude_paths(&blk)
        super do |exclude_paths|
          exclude_paths << '.dapp_chef'

          yield exclude_paths if block_given?
        end
      end

      def builder_cookbook_path
        Pathname.new(path).join('.dapp_chef')
      end
    end # Chef
  end # Dapp
end # Dapp
