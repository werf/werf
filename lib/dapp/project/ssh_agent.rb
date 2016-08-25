module Dapp
  # Project
  class Project
    # SshAgent
    module SshAgent
      class << self
        def included(_base)
          ::Dapp::Helper::Shellout.default_env_keys << 'SSH_AUTH_SOCK'
        end
      end # << self

      def run_ssh_agent
        sock_name = "dapp-ssh-#{SecureRandom.uuid}"

        "/tmp/#{sock_name}".tap do |sock_path|
          Process.fork do
            Prctl.call(Prctl::PR_SET_PDEATHSIG, Signal.list['TERM'], 0, 0, 0)

            Process.setproctitle sock_name

            @ssh_agent_pid = nil

            Signal.trap('INT') {  }
            Signal.trap('TERM') { Process.kill('TERM', @ssh_agent_pid) if @ssh_agent_pid }

            @ssh_agent_pid =  Process.fork do
              STDOUT.reopen '/dev/null', 'a'
              STDERR.reopen '/dev/null', 'a'
              exec 'ssh-agent', '-d', '-a', sock_path
            end

            Process.wait @ssh_agent_pid
          end

          begin
            ::Timeout.timeout(10) do
              until File.exist? sock_path
                sleep 0.001
              end
            end
          rescue ::Timeout::Error
            raise ::Dapp::Error::Project, code: :cannot_run_ssh_agent
          end
        end # sock_path
      end

      def add_ssh_key(ssh_key_path, **kwargs)
        shellout! "ssh-add #{ssh_key_path}", env: {SSH_AUTH_SOCK: ssh_auth_sock(**kwargs)}
      end

      def ssh_auth_sock(force_run_agent: false)
        @ssh_auth_sock ||= begin
          system_ssh_auth_sock = nil
          system_ssh_auth_sock = File.expand_path(ENV['SSH_AUTH_SOCK']) if ENV['SSH_AUTH_SOCK'] && File.exist?(ENV['SSH_AUTH_SOCK'])

          if force_run_agent || !system_ssh_auth_sock
            run_ssh_agent.tap { |ssh_auth_sock| ENV['SSH_AUTH_SOCK'] = ssh_auth_sock }
          else
            system_ssh_auth_sock
          end
        end
      end

      def setup_ssh_agent
        return unless cli_options[:ssh_key]

        cli_options[:ssh_key].each do |ssh_key|
          raise ::Dapp::Error::Project, code: :ssh_key_not_found,
                                        data: { path: ssh_key } unless File.exist? ssh_key

          File.chmod 0600, ssh_key
          add_ssh_key ssh_key, force_run_agent: true
        end
      end
    end # SshAgent
  end # Project
end # Dapp
