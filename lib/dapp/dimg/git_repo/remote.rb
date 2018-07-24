module Dapp
  module Dimg
    module GitRepo
      class Remote < Base
        CACHE_VERSION = 3

        attr_reader :url

        class << self
          def get_or_create(dapp, name, url:)
            repositories[url] ||= new(dapp, name, url: url).tap(&:clone_and_fetch)
          end

          def repositories
            @repositories ||= {}
          end
        end

        def initialize(dapp, name, url:)
          super(dapp, name)

          @url = url
        end

        def ruby2go_method(method, args_hash={})
          res = dapp.ruby2go_git_repo(args_hash.merge("RemoteGitRepo" => JSON.dump(get_ruby2go_state_hash), "method" => method))
          self.set_ruby2go_state_hash(JSON.load(res["data"]["state"]))

          if res["error"]
            raise res["error"]
          else
            res["data"]["result"]
          end
        end

        def get_ruby2go_state_hash
          super.tap {|res|
            res["Url"] = @url.to_s
            res["ClonePath"] = dapp.build_path("remote_git_repo", CACHE_VERSION.to_s, dapp.consistent_uniq_slugify(name), remote_origin_url_protocol).to_s # FIXME
            res["IsDryRun"] = dapp.dry_run?
          }
        end

        def set_ruby2go_state_hash(state)
          super(state)
          @url = state["Url"]
        end

        def _with_lock(&blk)
          dapp.lock("remote_git_artifact.#{name}", default_timeout: 600, &blk)
        end

        def _rugged_credentials
          @_rugged_credentials ||= begin
            if remote_origin_url_protocol == :ssh
              host_with_user = url.split(':', 2).first
              username = host_with_user.split('@', 2).reverse.last
              Rugged::Credentials::SshKeyFromAgent.new(username: username)
            end
          end
        end

        def path
          Pathname(dapp.build_path("remote_git_repo", CACHE_VERSION.to_s, dapp.consistent_uniq_slugify(name), remote_origin_url_protocol).to_s)
        end

        def clone_and_fetch
          return ruby2go_method("CloneAndFetch")
        end

        def latest_branch_commit(branch)
          git.ref("refs/remotes/#{branch_format(branch)}").tap do |ref|
            raise Error::Rugged, code: :branch_not_exist_in_remote_git_repository, data: { branch: branch, url: url } if ref.nil?
            break ref.target_id
          end
        end

        def latest_tag_commit(tag)
          git.tags[tag].tap do |t|
            raise Error::Rugged, code: :tag_not_exist_in_remote_git_repository, data: { tag: tag, url: url } if t.nil?
            break t.target_id
          end
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError, TypeError => _e
          raise Error::Rugged, code: :commit_not_found_in_remote_git_repository, data: { commit: commit, url: url }
        end

        def raise_submodule_commit_not_found!(commit)
          raise Error::Rugged, code: :git_remote_submodule_commit_not_found, data: { commit: commit, url: url }
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
