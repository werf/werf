module Dapp
  module Dimg
    module GitRepo
      class Remote < Base
        attr_reader :url

        def initialize(dimg, name, url:)
          super(dimg, name)

          @url = url

          dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.git_artifact_clone', data: { url: url }), short: true) do
            begin
              Rugged::Repository.clone_at(url, path.to_s, bare: true, credentials: _rugged_credentials)
            rescue Rugged::NetworkError, Rugged::SslError => e
              raise Error::Rugged, code: :rugged_remote_error, data: { message: e.message, url: url }
            end
          end unless path.directory?
        end

        def _rugged_credentials
          @_rugged_credentials ||= begin
            ssh_url = begin
              URI.parse(url)
              false
            rescue URI::InvalidURIError
              true
            end

            if ssh_url
              host_with_user = url.split(':', 2).first
              username = host_with_user.split('@', 2).reverse.last
              Rugged::Credentials::SshKeyFromAgent.new(username: username)
            end
          end
        end

        def path
          Pathname(dimg.build_path('git_repo_remote', name, Digest::MD5.hexdigest(url)).to_s)
        end

        def fetch!(branch = nil)
          branch ||= self.branch
          dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.git_artifact_fetch', data: { url: url }), short: true) do
            git.fetch('origin', [branch], credentials: _rugged_credentials)
            raise Error::Rugged, code: :branch_not_exist_in_remote_git_repository, data: { branch: branch, url: url } unless branch_exist?(branch)
          end unless dimg.ignore_git_fetch || dimg.dapp.dry_run?
        end

        def branch_exist?(name)
          git.branches.exist?(branch_format(name))
        end

        def latest_commit(name)
          git.ref("refs/remotes/#{branch_format(name)}").target_id
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError, TypeError => _e
          raise Error::Rugged, code: :commit_not_found_in_remote_git_repository, data: { commit: commit, url: url }
        end

        protected

        def git
          super(bare: true, credentials: _rugged_credentials)
        end

        private

        def branch_format(name)
          "origin/#{name.reverse.chomp('origin/'.reverse).reverse}"
        end
      end
    end
  end
end
