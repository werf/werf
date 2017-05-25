module Dapp
  # Project
  class Project
    # Chef
    module Chef
      def local_git_artifact_exclude_paths(&blk)
        super do |exclude_paths|
          exclude_paths << '.dapp_chef'
          exclude_paths << '.chefinit'

          yield exclude_paths if block_given?
        end
      end

      def local_cookbook_path
        File.join(path, '.dapp_chef')
      end

      def chefinit_cookbook_path
        File.join(path, '.chefinit')
      end
    end # Chef
  end # Project
end # Dapp
