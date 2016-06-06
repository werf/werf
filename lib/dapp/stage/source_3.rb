module Dapp
  module Stage
    class Source3 < Base
      def image
        super do |image|
          builder.git_artifact_list.each do |git_artifact|
            git_artifact.apply_source_3!(image)
          end
        end
      end

      def signature
        hashsum [builder.stages[:infra_setup].signature,
                 app_setup_file,
                 *builder.app_setup_commands, # TODO chef
                 *builder.git_artifact_list.map { |git_artifact| git_artifact.source_3_commit }]
      end

      def app_setup_file
        @app_setup_file ||= begin
          File.read(app_setup_file_path) if app_setup_file?
        end
      end

      def app_setup_file?
        File.exist?(app_setup_file_path)
      end

      def app_setup_file_path
        builder.build_path('.app_setup')
      end
    end # Source3
  end # Stage
end # Dapp
