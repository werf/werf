module Dapp
  module Dimg
    module GitRepo
      # Normal Git repo
      class Remote < Base
        def initialize(dimg, name, url:)
          super(dimg, name)

          @url = url

          dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.git_artifact_clone', data: { name: name }), short: true) do
            begin
              Rugged::Repository.clone_at(url, path, bare: true)
            rescue Rugged::NetworkError, Rugged::SslError => e
              raise Error::Rugged, code: :rugged_remote_error, data: { message: e.message, url: url }
            end
          end unless File.directory?(path)
        end

        def path
          dimg.build_path("#{name}.git").to_s
        end

        def fetch!(branch = nil)
          branch ||= self.branch
          dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.git_artifact_fetch', data: { name: name }), short: true) do
            git.fetch('origin', [branch])
          end unless dimg.ignore_git_fetch || dimg.dapp.dry_run?
        end

        def latest_commit(branch)
          git.ref("refs/remotes/origin/#{branch}").target_id
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError => _e
          raise Error::Rugged, code: :commit_not_found_in_remote_git_repository, data: { commit: commit, url: url }
        end

        protected

        attr_reader :url

        def git
          super(bare: true)
        end
      end
    end
  end
end
