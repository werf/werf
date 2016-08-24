module Dapp
  # Project
  class Project
    # SshAgent
    module SshAgent
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

      def ssh_auth_sock
        @ssh_auth_sock ||= begin
          if cli_options[:ssh_key]
            run_ssh_agent.tap do |ssh_auth_sock|
              ENV['SSH_AUTH_SOCK'] = ssh_auth_sock
              cli_options[:ssh_key].each { |ssh_key| shellout! "ssh-add #{ssh_key}" }
            end
          elsif ENV['SSH_AUTH_SOCK'] && File.exist?(ENV['SSH_AUTH_SOCK'])
            File.expand_path(ENV['SSH_AUTH_SOCK'])
          end
        end
      end

      def setup_ssh_agent
        ssh_auth_sock
      end
    end # SshAgent
  end # Project
end # Dapp
