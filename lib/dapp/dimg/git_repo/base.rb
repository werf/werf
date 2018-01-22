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

        def remote_origin_url
          @remote_origin_url ||= begin
            ro_url = url
            while url_protocol(ro_url) == :noname
              begin
                parent_git = Rugged::Repository.discover(ro_url)
              rescue Rugged::OSError
                parent_git_path = parent_git ? parent_git.path : path
                raise Error::Rugged, code: :git_repository_not_found, data: { path: ro_url, parent_git_path: parent_git_path }
              end

              ro_url = begin
                git_url(parent_git)
              rescue Error::Rugged => e
                break if e.net_status[:code] == :git_repository_without_remote_url # local repository
                raise
              end
            end

            ro_url
          end
        end

        def remote_origin_url_protocol
          url_protocol(remote_origin_url)
        end

        def nested_git_directories_patches(*_args)
          raise
        end

        def submodules_params(commit, paths: [], exclude_paths: [])
          raise "Workdir not supported for #{self.class}" if commit.nil?

          entry = begin
            lookup_commit(commit).tree.path('.gitmodules')
          rescue Rugged::TreeError
            return []
          end

          submodules_params_base(lookup_object(entry[:oid]).content, paths: paths, exclude_paths: exclude_paths)
        end

        def submodules_params_base(gitsubmodule_content, paths: [], exclude_paths: [])
          IniFile.new.parse(gitsubmodule_content)
            .to_h
            .select { |_, params| !ignore_directory?(params['path'], paths: paths, exclude_paths: exclude_paths) }
            .map do |_, params|
              params = params.symbolize_keys
              params[:branch] = params[:branch].to_s if params.key?(:branch)
              params[:url] = submodule_url(params[:url])
              params[:type] = url_protocol(params[:url]) == :noname ? :local : :remote
              params
            end
        end

        def submodule_url(gitsubmodule_url)
          if gitsubmodule_url.start_with?('../')
            case remote_origin_url_protocol
            when :http, :https, :git
              uri = URI.parse(remote_origin_url)
              uri.path = File.expand_path(File.join(uri.path, gitsubmodule_url))
              uri.to_s
            when :ssh
              host_with_user, group_and_project = remote_origin_url.split(':', 2)
              group_and_project = File.expand_path(File.join('/', group_and_project, gitsubmodule_url))[1..-1]
              [host_with_user, group_and_project].join(':')
            else
              raise
            end
          else
            gitsubmodule_url
          end
        end

        # FIXME: Убрать логику исключения путей exclude_paths из данного класса,
        # FIXME: т.к. большинство методов не поддерживают инвариант
        # FIXME "всегда выдавать данные с исключенными путями".
        # FIXME: Например, метод diff выдает данные без учета exclude_paths.
        # FIXME: Лучше перенести фильтрацию в GitArtifact::diff_patches.
        # FIXME: ИЛИ обеспечить этот инвариант, но это ограничит в возможностях
        # FIXME: использование Rugged извне этого класса и это более сложный путь.
        # FIXME: Лучше сейчас убрать фильтрацию, а добавить ее когда наберется достаточно
        # FIXME: примеров использования.

        def patches(from, to, paths: [], exclude_paths: [], **kwargs)
          diff(from, to, **kwargs).patches.select do |patch|
            ignore_patch?(patch, paths: paths, exclude_paths: exclude_paths)
          end
        end

        def ignore_patch?(patch, paths: [], exclude_paths: [])
          !ignore_path?(patch.delta.new_file[:path], paths: paths, exclude_paths: exclude_paths)
        end

        def entries(commit, paths: [], exclude_paths: [])
          [].tap do |entries|
            lookup_commit(commit).tree.walk(:preorder) do |root, entry|
              fullpath = File.join(root, entry[:name]).reverse.chomp('/').reverse
              next if entry[:type] == :tree || ignore_path?(fullpath, paths: paths, exclude_paths: exclude_paths)
              entries << [root, entry]
            end
          end
        end

        def diff(from, to, **kwargs)
          if to.nil?
            raise "Workdir diff not supported for #{self.class}"
          elsif from.nil?
            Rugged::Tree.diff(git, nil, to, **kwargs)
          else
            lookup_commit(from).diff(lookup_commit(to), **kwargs)
          end
        end

        def commit_exists?(commit)
          git.exists?(commit)
        end

        def latest_commit(_branch)
          raise
        end

        def branch
          git.head.name.sub(/^refs\/heads\//, '')
        rescue Rugged::ReferenceError => e
          raise Error::Rugged, code: :git_repository_reference_error, data: { name: name, message: e.message.downcase }
        end

        def tag_at(name)
          tag = git.tags.find { |t| t.name == name }
          tag.target.time.to_i
        end

        def tags
          git.tags.map(&:name)
        end

        def remote_branches
          git.branches
            .map(&:name)
            .select { |b| b.start_with?('origin/') }
            .map { |b| b.reverse.chomp('origin/'.reverse).reverse }
        end

        def commit_at(commit)
          lookup_commit(commit).time.to_i
        end

        def find_commit_id_by_message(regex)
          walker.each do |commit|
            msg = commit.message.encode('UTF-8', invalid: :replace, undef: :replace)
            return commit.oid if msg =~ regex
          end
        end

        def walker
          walker = Rugged::Walker.new(git)
          walker.push(git.head.target_id)
          walker
        end

        def lookup_object(oid)
          git.lookup(oid)
        end

        def lookup_commit(commit)
          git.lookup(commit)
        end

        protected

        def git(**kwargs)
          @git ||= Rugged::Repository.new(path.to_s, **kwargs)
        end

        def url
          @url ||= git_config_remote_origin_url(git)
        end

        def git_url(git_repo)
          git_config_remote_origin_url(git_repo)
        end

        def git_config_remote_origin_url(git_repo)
          git_repo.config.to_hash['remote.origin.url'].tap do |url|
            raise Error::Rugged, code: :git_repository_without_remote_url, data: { name: self.class, path: git_repo.path } if url.nil?
          end
        end

        def url_protocol(url)
          if (scheme = URI.parse(url).scheme).nil?
            :noname
          else
            scheme.to_sym
          end
        rescue URI::InvalidURIError
          :ssh
        rescue Error::Rugged => e
          return :none if e.net_status[:code] == :git_repository_without_remote_url
          raise
        end

        private

        def ignore_directory?(path, paths: [], exclude_paths: [])
          ignore_path_base(path, exclude_paths: exclude_paths) do
            paths.empty? || paths.any? { |p| check_path?(path, p) || check_subpath?(path, p) }
          end
        end

        def ignore_path?(path, paths: [], exclude_paths: [])
          ignore_path_base(path, exclude_paths: exclude_paths) do
            paths.empty? ||
              paths.any? do |p|
                File.fnmatch?(p, path, File::FNM_PATHNAME) ||
                  File.fnmatch?(File.join(p, '**', '*'), path, File::FNM_PATHNAME)
              end
          end
        end

        def ignore_path_base(path, exclude_paths: [])
          is_exclude_path = exclude_paths.any? { |p| check_path?(path, p) }
          is_include_path = yield
          is_exclude_path || !is_include_path
        end
      end
    end
  end
end
