module Dapper
  module GitRepo
    class Remote < Base
      def initialize(builder, name, url:, ssh_key_path: nil, **kwargs)
        super(builder, name, **kwargs)

        @url = url
        @ssh_key_path = File.expand_path(ssh_key_path, builder.home_path) if ssh_key_path

        @use_ssh_key = false
        File.chmod(0600, @ssh_key_path) if @ssh_key_path

        lock do
          unless File.directory? dir_path
            with_ssh_key do
              git "clone --bare --depth 1 #{url} #{dir_path}", log_verbose: true
            end
          end
        end
      end

      def fetch!(branch = 'master')
        lock do
          with_ssh_key do
            git_bare "fetch origin #{branch}:#{branch}", log_verbose: true
          end
        end
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
          builder.shellout "ssh-agent bash -ec 'ssh-add #{ssh_key_path}; git #{command}'", **kwargs
        else
          super
        end
      end
    end
  end
end
