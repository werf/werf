module Dapp
  module Dimg
    module GitRepo
      class Local < Base
        attr_reader :path

        def initialize(dapp, name, path)
          super(dapp, name)
          self.path = path
        end

        def path=(path)
          @path ||= Pathname(Rugged::Repository.new(path).path)
        rescue Rugged::RepositoryError, Rugged::OSError => _e
          raise Error::Rugged, code: :local_git_repository_does_not_exist, data: { path: path }
        end

        def workdir_path
          Pathname(git.workdir)
        end

        def nested_git_directories_patches(paths: [], exclude_paths: [], **kwargs)
          patches(nil, nil, paths: paths, exclude_paths: exclude_paths, **kwargs).select do |patch|
            delta_new_file = patch.delta.new_file
            nested_git_repository_mode?(delta_new_file[:mode])
          end
        end

        def submodules_params(commit, paths: [], exclude_paths: [])
          submodules(commit, paths: paths, exclude_paths: exclude_paths).map do |submodule|
            next if commit.nil? && !submodule.in_config?
            submodule_params(submodule).tap do |params|
              params[:commit] = submodule.workdir_oid || params[:commit] if commit.nil?
              if submodule.in_workdir? && !submodule.uninitialized?
                dapp.log_info("Using local submodule repository `#{params[:path]}`!")
                params[:type] = :local
              end
            end
          end.compact
        end

        def raise_submodule_commit_not_found!(commit)
          raise Error::Rugged, code: :git_local_submodule_commit_not_found, data: { commit: commit, path: path }
        end

        def ignore_patch?(patch, paths: [], exclude_paths: [])
          delta_new_file = patch.delta.new_file
          args = [delta_new_file[:path], paths: paths, exclude_paths: exclude_paths]
          if nested_git_repository_mode?(delta_new_file[:mode])
            ignore_directory?(*args)
          else
            ignore_path?(*args)
          end
        end

        def nested_git_repository_mode?(mode)
          mode == 0o040000
        end

        # NOTICE: Параметры {from: nil, to: nil} можно указать только для Own repo.
        # NOTICE: Для Remote repo такой вызов не имеет смысла и это ошибка пользователя класса Remote.

        def diff(from, to, **kwargs)
          if from.nil? and to.nil?
            mid_commit = latest_commit
            diff_obj = super(nil, mid_commit, **kwargs)
            diff_obj.merge! git.lookup(mid_commit).diff_workdir(**kwargs)
            diff_obj
          elsif to.nil?
            git.lookup(from).diff_workdir(**kwargs)
          else
            super
          end
        end

        def latest_commit(_branch = nil)
          git.head.target_id
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError, TypeError => _e
          raise Error::Rugged, code: :commit_not_found_in_local_git_repository, data: { commit: commit, path: path }
        end

        protected

        def git_repo_exist?(path)
          Rugged::Repository.new(path)
          true
        rescue Rugged::RepositoryError, Rugged::OSError => _e
          false
        end
      end
    end
  end
end
