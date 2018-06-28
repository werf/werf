module Dapp
  class Dapp
    module Command
      module Sample
        module Common
          def _sample_list
            @sample_list ||= [].tap do |list|
              tree = begin
                latest_commit = _sample_git_repo.latest_commit(_samples_git_repo_branch)
                latest_commit_tree = _sample_git_repo.lookup_commit(latest_commit).tree

                if _samples_dir == '.'
                  latest_commit_tree
                else
                  begin
                    oid = latest_commit_tree.path(_samples_dir)[:oid]
                  rescue Rugged::TreeError
                    raise Error::Command, code: :samples_directory_not_exist, data: { url: _samples_git_repo_url, path: _samples_dir }
                  end

                  _sample_git_repo.lookup_object(oid)
                end
              end

              tree.each_tree { |entry| list << entry[:name] }
            end
          end

          def _sample_git_repo
            @sample_repo ||= begin
              Dimg::GitRepo::Remote.get_or_create(
                self,
                git_url_to_name(_samples_git_repo_url),
                url: _samples_git_repo_url
              )
            end
          end

          def _samples_git_repo_url
            options[:samples_repo]
          end

          def _samples_git_repo_branch
            options[:samples_branch]
          end

          def _samples_dir
            options[:samples_dir]
          end
        end
      end
    end
  end
end
