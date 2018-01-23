module Dapp
  module Dimg
    module Config
      module Directive
        class Dimg < Base
          module InstanceMethods
            attr_reader :_builder
            attr_reader :_chef, :_shell, :_docker, :_git_artifact, :_mount, :_artifact
            attr_reader :_artifact_groups

            def chef(&blk)
              builder(:chef)
              directive_eval(_chef, &blk)
            end

            def shell(&blk)
              builder(:shell)
              directive_eval(_shell, &blk)
            end

            def docker(&blk)
              directive_eval(_docker, &blk)
            end

            def artifact(name = nil, &blk)
              pass_to(ArtifactGroup.new(name, dapp: dapp), :clone_to_artifact).tap do |artifact_group|
                _context_artifact_groups << directive_eval(artifact_group, &blk)
                dapp.artifact_config(name, artifact_group._artifact_config) unless name.nil?
              end
            end

            def import(name, from, &blk)
              ArtifactGroup.new(dapp: dapp).tap do |artifact_group|
                artifact_group._artifact_export(dapp.artifact_config_by_name(name), from, &blk)
                _context_artifact_groups << artifact_group
              end
            end

            def git(url = nil, &blk)
              type = url.nil? ? :local : :remote
              _git_artifact.public_send(type, url.to_s, &blk)
            end

            def mount(to, &blk)
              Mount.new(to, dapp: dapp, &blk).tap do |mount|
                _mount << mount
              end
            end

            def _builder
              @_builder || :none
            end

            def _chef
              @_chef ||= Chef.new(dapp: dapp)
            end

            def _shell
              @_shell ||= Shell::Dimg.new(dapp: dapp)
            end

            def _docker
              @_docker ||= Docker::Dimg.new(dapp: dapp)
            end

            def _mount
              @_mount ||= []
            end

            def _git_artifact
              @_git_artifact ||= GitArtifact.new(dapp: dapp)
            end

            [:build_dir, :tmp_dir, :custom_dir].each do |mount_type|
              define_method "_#{mount_type}_mount" do
                _mount.select { |m| m._type == mount_type }
              end
            end

            def _artifact_groups
              @_artifact_groups ||= []
            end

            def _context_artifact_groups
              @_context_artifact_groups ||= []
            end

            def _artifact
              [_artifact_groups, _context_artifact_groups].flatten.map(&:_export).flatten
            end

            [:before, :after].each do |order|
              [:setup, :install].each do |stage|
                define_method "_#{order}_#{stage}_artifact" do
                  _artifact.select do |art|
                    art.public_send("_#{order}") == stage
                  end
                end
              end
            end

            def _import_artifact
              _artifact.select(&:not_associated?)
            end

            class GitArtifact < Base
              attr_reader :_local, :_remote

              def initialize(**kwargs, &blk)
                @_local = []
                @_remote = []

                super(**kwargs, &blk)
              end

              def local(_, &blk)
                GitArtifactLocal.new(dapp: dapp, &blk).tap do |git|
                  @_local << git
                end
              end

              def remote(repo_url, &blk)
                GitArtifactRemote.new(repo_url, dapp: dapp, &blk).tap do |git|
                  @_remote << git
                end
              end

              def _local
                @_local.map(&:_export).flatten
              end

              def _remote
                @_remote.map(&:_export).flatten
              end

              def empty?
                (_local + _remote).empty?
              end

              def validate!
                (_local + _remote).each(&:validate!)
              end
            end

            def artifacts_after_parsing!
              _artifacts_auto_excluding!
              _artifact.map(&:_config).each(&:artifacts_after_parsing!)
            end

            protected

            def builder(type)
              @_builder = type if _builder == :none
              raise ::Dapp::Error::Config, code: :builder_type_conflict unless @_builder == type
            end

            def passed_directives
              [:@_chef, :@_shell, :@_docker, :@_git_artifact, :@_mount, :@_artifact_groups, :@_builder]
            end

            def _artifacts_auto_excluding!
              path_to_relative = proc { |path| path.reverse.chomp('/').reverse }

              all_artifacts.reduce({}) do |hash, artifact|
                unless artifact._to.nil?
                  to_common = artifact._to[/^\/[^\/]*/]
                  hash[to_common] ||= []
                  hash[to_common] << artifact
                end
                hash
              end.each do |to_common, artifacts|
                include_paths_common = artifacts.reduce([]) do |arr, artifact|
                  arr << artifact._to.sub(to_common, '')
                  arr.concat(artifact._include_paths.map { |path| File.join(artifact._to.sub(to_common, ''), path) } )
                  arr
                end.map(&path_to_relative).uniq

                artifacts.each do |artifact|
                  artifact_include_shift = path_to_relative.call(artifact._to.sub(to_common, ''))

                  include_paths_common.each do |path|
                    next if artifact_include_shift.start_with? path

                    path = path_to_relative.call(path.sub(artifact_include_shift, ''))
                    unless artifact._include_paths.any? { |ipath| ipath.start_with? path } || artifact._exclude_paths.any? { |epath| path.start_with? epath }
                      artifact._exclude_paths << path
                    end
                  end
                end
              end
            end

            def all_artifacts
              _artifact + _git_artifact._local + _git_artifact._remote
            end
          end # InstanceMethods
          # rubocop:enable Metrics/ModuleLength
        end # Dimg
      end # Directive
    end # Config
  end # Dimg
end # Dapp
