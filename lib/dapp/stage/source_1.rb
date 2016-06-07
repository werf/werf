module Dapp
  module Stage
    class Source1 < Base
      def name
        :source_1
      end

      def prev_source_stage_name
        :source_1_archive
      end

      def image
        super do |image|
          build.git_artifact_list.each do |git_artifact|
            git_artifact.source_1_apply!(image)
          end
        end
      end

      def signature
        hashsum [build.stages[:source_1_archive].signature,
                 dependency_file, dependency_file_regex,
                 *build.app_install_commands, # TODO chef
                 *build.git_artifact_list.map { |git_artifact| git_artifact.source_1_commit }]
      end

      def git_artifact_signature
        hashsum [build.stages[:source_1_archive].signature,
                 *build.app_install_commands]
      end

      def dependency_file
        @dependency_file ||= begin
          file_path = Dir[build.build_path('*')].detect {|x| x =~ dependency_file_regex }
          File.read(file_path) unless file_path.nil?
        end
      end

      def dependency_file?
        !dependency_file.nil?
      end

      def dependency_file_regex
        /.*\/(Gemfile|composer.json|requirement_file.txt)$/
      end
    end # Source1
  end # Stage
end # Dapp
