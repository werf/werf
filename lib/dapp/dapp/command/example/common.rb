module Dapp
  class Dapp
    module Command
      module Example
        module Common
          def _example_list
            @example_list ||= [].tap do |list|
              tree = begin
                latest_commit = _example_git_repo.latest_commit(_examples_git_repo_branch)
                latest_commit_tree = _example_git_repo.lookup_commit(latest_commit).tree

                if _examples_dir == '.'
                  latest_commit_tree
                else
                  begin
                    oid = latest_commit_tree.path(_examples_dir)[:oid]
                  rescue Rugged::TreeError
                    raise Error::Command, code: :examples_directory_not_exist, data: { url: _examples_git_repo_url, path: _examples_dir }
                  end

                  _example_git_repo.lookup_object(oid)
                end
              end

              tree.each_tree { |entry| list << entry[:name] }
            end
          end

          def _example_git_repo
            @example_repo ||= begin
              Dimg::GitRepo::Remote.get_or_create(
                self,
                git_url_to_name(_examples_git_repo_url),
                url: _examples_git_repo_url
              )
            end
          end

          def _examples_git_repo_url
            options[:examples_repo]
          end

          def _examples_git_repo_branch
            options[:examples_branch]
          end

          def _examples_dir
            options[:examples_dir]
          end
        end
      end
    end
  end
end
