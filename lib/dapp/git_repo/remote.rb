module Dapp
  module GitRepo
    # Normal Git repo
    class Remote < Base
      def initialize(dimg, name, url:)
        super(dimg, name)

        @url = url

        dimg.project.log_secondary_process(dimg.project.t(code: 'process.git_artifact_clone', data: { name: name }), short: true) do
          git "clone --bare --depth 1 #{url} #{path}"
        end unless File.directory?(path)
      end

      def fetch!(branch = 'master')
        dimg.project.log_secondary_process(dimg.project.t(code: 'process.git_artifact_fetch', data: { name: name }), short: true) do
          git_bare "fetch origin #{branch}:#{branch}"
        end unless dimg.ignore_git_fetch || dimg.project.dry_run?
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
