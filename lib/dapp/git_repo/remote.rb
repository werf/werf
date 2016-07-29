module Dapp
  module GitRepo
    # Normal Git repo
    class Remote < Base
      def initialize(application, name, url:, ssh_key_path: nil)
        super(application, name)

        @url = url
        @ssh_key_path = File.expand_path(ssh_key_path, application.home_path) if ssh_key_path

        @use_ssh_key = false
        File.chmod(0o600, @ssh_key_path) if @ssh_key_path

        with_ssh_key do
          git "clone --bare --depth 1 #{url} #{path}"
        end unless File.directory?(path)
      end

      def fetch!(branch = 'master')
        with_ssh_key do
          application.log_secondary_process(application.t(code: 'process.git_artifact_fetch', data: { name: name }), short: true) do
            git_bare "fetch origin #{branch}:#{branch}"
          end
        end unless application.ignore_git_fetch || application.dry_run?
      end

      def cleanup!
        super
        FileUtils.rm_rf path
      end

      protected

      attr_reader :url
      attr_reader :ssh_key_path

      attr_accessor :use_ssh_key
      def use_ssh_key
        @use_ssh_key ||= false
      end

      def with_ssh_key
        original = use_ssh_key
        self.use_ssh_key = true

        yield
      ensure
        self.use_ssh_key = original
      end

      def git(command, **kwargs)
        if use_ssh_key && ssh_key_path
          application.shellout!("ssh-agent bash -ec 'ssh-add #{ssh_key_path}; git #{command}'", **kwargs)
        else
          super
        end
      end
    end
  end
end
