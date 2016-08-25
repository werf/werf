module Dapp
  module GitRepo
    # Normal Git repo
    class Remote < Base
      def initialize(application, name, url:)
        super(application, name)

        @url = url

        application.project.log_secondary_process(application.project.t(code: 'process.git_artifact_clone', data: { name: name }), short: true) do
          git "clone --bare --depth 1 #{url} #{path}"
        end unless File.directory?(path)
      end

      def fetch!(branch = 'master')
        application.project.log_secondary_process(application.project.t(code: 'process.git_artifact_fetch', data: { name: name }), short: true) do
          git_bare "fetch origin #{branch}:#{branch}"
        end unless application.ignore_git_fetch || application.project.dry_run?
      end

      def cleanup!
        super
        FileUtils.rm_rf path
      end

      protected

      attr_reader :url
    end
  end
end
