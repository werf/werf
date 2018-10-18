module Dapp
  module Dimg
    module GitRepo
      class Remote < Base
        CACHE_VERSION = 4

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

        def url
          @url ||= remote_origin_url
        end

        def ruby2go_method(method, args_hash={})
          res = dapp.ruby2go_git_repo(args_hash.merge("RemoteGitRepo" => JSON.dump(get_ruby2go_state_hash), "method" => method))

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["RemoteGitRepo"]))

          res["data"]["result"]
        end

        def get_ruby2go_state_hash
          super.tap {|res|
            res["Url"] = @url.to_s
            res["ClonePath"] = dapp.build_path("remote_git_repo", CACHE_VERSION.to_s, dapp.consistent_uniq_slugify(name), url_protocol(url)).to_s # FIXME
            res["IsDryRun"] = dapp.dry_run?
          }
        end

        def set_ruby2go_state_hash(state)
          super(state)
          @url = state["Url"]
        end

        def _rugged_credentials # TODO remove
          @_rugged_credentials ||= begin
            if url_protocol(url) == :ssh
              host_with_user = url.split(':', 2).first
              username = host_with_user.split('@', 2).reverse.last
              Rugged::Credentials::SshKeyFromAgent.new(username: username)
            end
          end
        end

        def path
          Pathname(dapp.build_path("remote_git_repo", CACHE_VERSION.to_s, dapp.consistent_uniq_slugify(name), url_protocol(url)).to_s)
        end

        def clone_and_fetch
          return ruby2go_method("CloneAndFetch")
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError, TypeError => _e
          raise Error::Rugged, code: :commit_not_found_in_remote_git_repository, data: { commit: commit, url: url }
        end

        protected

        def git # TODO remove
          super(bare: true, credentials: _rugged_credentials)
        end

        private

        def url_protocol(url) # TODO remove
          if (scheme = URI.parse(url).scheme).nil?
            :noname
          else
            scheme.to_sym
          end
        rescue URI::InvalidURIError
          :ssh
        end

        def branch_format(name) # TODO remove
          "origin/#{name.reverse.chomp('origin/'.reverse).reverse}"
        end
      end
    end
  end
end
