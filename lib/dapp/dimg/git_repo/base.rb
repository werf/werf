module Dapp
  module Dimg
    module GitRepo
      # Base class for any Git repo (remote, gitkeeper, etc)
      class Base
        include Helper::Trivia

        attr_reader :name
        attr_reader :dapp

        def initialize(dapp, name)
          @dapp = dapp
          @name = name
        end

        def exclude_paths
          []
        end

        def get_ruby2go_state_hash
          {
            "Name" => @name.to_s,
          }
        end

        def set_ruby2go_state_hash(state)
          @name = state["Name"]
        end

        def empty?
          return @empty if !@empty.nil?
          @empty = ruby2go_method("IsEmpty")
          @empty
        end

        def find_commit_id_by_message(regex)
          ruby2go_method("FindCommitIdByMessage", "regex" => regex)
        end

        def commit_exists?(commit)
          ruby2go_method("IsCommitExists", "commit" => commit)
        end

        def remote_origin_url
          return @remote_origin_url if @remote_origin_url_set

          res = ruby2go_method("RemoteOriginUrl")
          if res != ""
            @remote_origin_url = res
          end
          @remote_origin_url_set = true

          @remote_origin_url
        end

        def latest_branch_commit(branch)
          ruby2go_method("LatestBranchCommit", "branch" => branch)
        end

        def head_commit
          return @head_commit if !@head_commit.nil?
          @head_commit = ruby2go_method("HeadCommit")
          @head_commit
        end

        def head_branch_name
          return @branch if !@branch.nil?
          @branch = ruby2go_method("HeadBranchName")
          @branch
        end

        # TODO Below operations does not affect build process, only used in sample.
        # TODO Should be ported to golang without git_repo.GitRepo interface.

        def lookup_object(oid)
          git.lookup(oid)
        end

        def lookup_commit(commit)
          git.lookup(commit)
        end

        protected

        def ruby2go_method(method, args_hash={})
          raise
        end

        def git(**kwargs) # TODO remove
          @git ||= Rugged::Repository.new(path.to_s, **kwargs)
        end

      end
    end
  end
end
