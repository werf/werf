module Dapper
  module GitRepo
    class Chronicler < Base
      def initialize(builder, name, **kwargs)
        super

        lock do
          unless File.directory? chronodir_path
            git "init --separate-git-dir=#{dir_path} #{chronodir_path}"
            git_chrono 'commit --allow-empty -m init'
          end
        end
      end

      def chronodir_path(*path)
        build_path name, *path
      end

      def commit!(comment = '+')
        lock do
          git_chrono 'add --all'
          unless git_chrono('diff --cached --quiet', returns: [0, 1]).status.success?
            git_chrono "commit -m #{comment}"
          end
        end
      end

      def cleanup!
        lock do
          super
          FileUtils.rm_rf chronodir_path
        end
      end

      protected

      def git_chrono(command, **kwargs)
        git "-C #{chronodir_path} #{command}", **kwargs
      end
    end
  end
end
