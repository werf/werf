module Dapp
  module Dimg
    module GitRepo
      class Local < Base
        attr_reader :path

        def initialize(dapp, name, path)
          super(dapp, name)
          self.path = path
        end

        def get_ruby2go_state_hash
          super.tap {|res|
            p = @path.to_s
            p = File.dirname(@path) if File.basename(@path) == ".git"
            res["Path"] = p
            res["GitDir"] = @path.to_s
          }
        end

        def ruby2go_method(method, args_hash={})
          res = dapp.ruby2go_git_repo(args_hash.merge("LocalGitRepo" => JSON.dump(get_ruby2go_state_hash), "method" => method))

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["LocalGitRepo"]))

          res["data"]["result"]
        end

        def path=(path) # TODO remove
          @path ||= Pathname(Rugged::Repository.new(path).path)
        rescue Rugged::RepositoryError, Rugged::OSError => _e
          raise Error::Rugged, code: :local_git_repository_does_not_exist, data: { path: path }
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError, TypeError => _e
          raise Error::Rugged, code: :commit_not_found_in_local_git_repository, data: { commit: commit, path: path }
        end
      end
    end
  end
end
