module Dapp
  module Config
    # Application
    class Application
      attr_reader :_home_path
      attr_reader :_basename
      attr_reader :_parent
      attr_reader :_builder, :_chef, :_shell
      attr_reader :_install_dependencies, :_setup_dependencies
      attr_reader :_docker
      attr_reader :_git_artifact
      attr_reader :_before_install_artifact, :_before_setup_artifact, :_after_install_artifact, :_after_setup_artifact, :_import_artifact
      attr_reader :_tmp_dir, :_build_dir

      def initialize(parent)
        @_parent = parent

        @_docker       = Directive::Docker::Base.new
        @_git_artifact = Directive::GitArtifact.new
        @_shell        = Directive::Shell::Base.new
        @_chef         = Directive::Chef.new
        @_tmp_dir      = Directive::TmpDir.new
        @_build_dir    = Directive::BuildDir.new

        @_apps                    = []

        @_before_install_artifact = []
        @_before_setup_artifact   = []
        @_after_install_artifact  = []
        @_after_setup_artifact    = []
        @_import_artifact         = []

        @_install_dependencies    = []
        @_setup_dependencies      = []

        yield self if block_given?
      end

      def install_depends_on(*args)
        @_install_dependencies.concat(args)
      end

      def setup_depends_on(*args)
        @_setup_dependencies.concat(args)
      end

      def builder(type)
        project.log_warning(desc: { code: 'excess_builder_instruction' }) if @_chef.send(:empty?) && @_shell.send(:empty?)
        raise Error::Config, code: :builder_type_unsupported, data: { type: type } unless [:chef, :shell].include?((type = type.to_sym))
        another_builder = [:chef, :shell].find { |t| t != type }
        instance_variable_set(:"@_#{another_builder}", instance_variable_get(:"@_#{another_builder}").class.new)
        @_builder = type
      end

      def chef
        @_chef.tap { raise Error::Config, code: :builder_type_conflict unless _builder == :chef }
      end

      def shell
        @_shell.tap { raise Error::Config, code: :builder_type_conflict unless _builder == :shell }
      end

      def docker
        @_docker
      end

      def artifact(where_to_add, before: nil, after: nil, **options, &blk)
        raise Error::Config, code: :stage_artifact_double_associate unless before.nil? || after.nil?
        artifact_base(instance_variable_get(:"@_#{artifact_variable_name(before, after)}"), where_to_add, **options, &blk)
      end

      def git_artifact
        @_git_artifact ||= Directive::GitArtifact.new
      end

      def tmp_dir
        @_tmp_dir
      end

      def build_dir
        @_build_dir
      end

      def dev_mode
        @_dev_mode = true
      end

      def _name
        (@_name || @_basename).tap do |name|
          reg = '^[[[:alnum:]]_.-]*$'
          raise Error::Config, code: :app_name_incorrect, data: { name: name, reg: reg } unless name =~ /#{reg}/
        end
      end

      def _dev_mode
        !!@_dev_mode
      end

      def _apps
        @_apps.empty? ? [self] : @_apps.flatten
      end

      def _app_chain
        @_app_chain ||= (_parent ? _parent._app_chain : []) + [self]
      end

      def _app_runlist
        _app_chain.map(&:_name).map do |name|
          if (subname = name.split("#{_root_app._name}-", 2)[1])
            subname_parts = subname.split('-')
            subname_parts.join('_') if subname_parts.any?
          end
        end.compact
      end

      def _root_app
        _app_chain.first
      end

      protected

      attr_accessor :project

      def clone_to_application
        clone_to(Application.new(self))
      end

      def clone_to_artifact
        clone_to(Artifact.new(self))
      end

      # rubocop:disable Metrics/AbcSize
      def clone_to(app)
        app.instance_variable_set(:'@project', project)
        app.instance_variable_set(:'@_builder', _builder)
        app.instance_variable_set(:'@_home_path', _home_path)
        app.instance_variable_set(:'@_basename', _basename)
        app.instance_variable_set(:'@_install_dependencies', _install_dependencies)
        app.instance_variable_set(:'@_setup_dependencies', _setup_dependencies)
        [:_before_install_artifact, :_before_setup_artifact, :_after_install_artifact, :_after_setup_artifact, :_import_artifact].each do |artifact|
          app.instance_variable_set(:"@#{artifact}", instance_variable_get(:"@#{artifact}").map { |a| a.send(:clone) })
        end
        app.instance_variable_set(:'@_docker', _docker.send(:clone))
        app.instance_variable_set(:'@_git_artifact', _git_artifact.send(:clone))
        app.instance_variable_set(:'@_chef', _chef.send(:clone))
        app.instance_variable_set(:'@_shell', _shell.send(:clone))
        app.instance_variable_set(:'@_tmp_dir', _tmp_dir.send(:clone))
        app.instance_variable_set(:'@_build_dir', _build_dir.send(:clone))
        app
      end
      # rubocop:enable Metrics/AbcSize

      def app(sub_name, &blk)
        clone_to_application.tap do |app|
          app.instance_variable_set(:'@_name', app_name(sub_name))
          app.instance_eval(&blk) if block_given?
          @_apps += app._apps
        end
      end

      def app_name(sub_name)
        [_name, sub_name].compact.join('-')
      end

      def artifact_variable_name(before, after)
        return :import_artifact if before.nil? && after.nil?

        if before.nil?
          prefix = :after
          stage  = after
        else
          prefix = :before
          stage  = before
        end

        return [prefix, stage, :artifact].join('_') if [:install, :setup].include?(stage.to_sym)
        raise Error::Config, code: :stage_artifact_not_supported_associated_stage, data: { stage: stage }
      end

      def artifact_base(artifact, where_to_add, **options, &blk)
        artifact << begin
          config = clone_to_artifact.tap do |app|
            app.instance_variable_set(:'@_shell', _shell.send(:clone_to_artifact))
            app.instance_variable_set(:'@_docker', _docker.send(:clone_to_artifact))
            app.instance_variable_set(:'@_name', app_name("artifact-#{SecureRandom.hex(2)}"))
            app.instance_eval(&blk) if block_given?
          end
          Directive::Artifact::Stage.new(where_to_add, config: config, **options)
        end
      end

      def validate!
        if _docker._from.nil?
          validate_scratch_directives!
          validate_scratch_artifacts!
        else
          raise Error::Config, code: :stage_artifact_not_associated unless _import_artifact.empty?
        end
        validate_artifacts!
        validate_artifacts_artifacts!
      end

      def validate_scratch_directives!
        raise Error::Config, code: :scratch_unsupported_directive, data: { directive: :app } unless _apps.length == 1

        directives = [:_shell, :_chef, :_git_artifact, :_install_dependencies, :_setup_dependencies, :_tmp_dir, :_build_dir]
        directives.each do |directive|
          raise Error::Config,
                code: :scratch_unsupported_directive,
                data: { directive: directive[1..-1] } unless public_send(directive).send(:empty?)
        end

        docker_directives = [:_expose, :_env, :_cmd, :_onbuild, :_workdir, :_user, :_entrypoint]
        docker_directives.each do |directive|
          value = _docker.public_send(directive)
          raise Error::Config,
                code: :scratch_unsupported_directive,
                data: { directive: "docker.#{directive[1..-1]}" } unless value.nil? || value.send(:empty?)
        end
      end

      def validate_scratch_artifacts!
        raise Error::Config, code: :scratch_artifact_associated unless associated_artifacts.empty?
        raise Error::Config, code: :scratch_artifact_required if _import_artifact.empty?
        _import_artifact.each do |artifact|
          raise Error::Config, code: :scratch_artifact_docker_from if artifact._config._docker._from.nil?
        end
      end

      def validate_artifacts_artifacts!
        associated_artifacts.each { |artifact| artifact._config.validate! }
      end

      def associated_artifacts
        _before_install_artifact + _before_setup_artifact + _after_install_artifact + _after_setup_artifact
      end

      def validate_artifacts!
        artifacts = validate_artifact_format(associated_artifacts + _import_artifact + _git_artifact._remote + _git_artifact._local)
        loop do
          break if artifacts.empty?
          verifiable_artifact = artifacts.shift
          artifacts.select { |artifact| artifact[:where_to_add] == verifiable_artifact[:where_to_add] }.each do |artifact|
            next if verifiable_artifact[:index] == artifact[:index]
            validate_artifact!(verifiable_artifact, artifact)
            validate_artifact!(artifact, verifiable_artifact)
          end
        end
      end

      def validate_artifact_format(artifacts)
        artifacts.map do |a|
          path_format = proc { |path| File.expand_path(File.join('/', path, '/'))[1..-1] }

          path_format.call(a._where_to_add) =~ %r{^([^\/]*)\/?(.*)$}
          where_to_add = Regexp.last_match(1)
          includes = a._paths.dup
          includes << Regexp.last_match(2) unless Regexp.last_match(2).empty?
          excludes = a._exclude_paths.dup

          {
            index: artifacts.index(a),
            where_to_add: where_to_add,
            includes: includes.map(&path_format),
            excludes: excludes.map(&path_format)
          }
        end
      end

      def validate_artifact!(verifiable_artifact, artifact)
        verifiable_artifact[:includes].each do |verifiable_path|
          potential_conflicts = artifact[:includes].select { |path| path.start_with?(verifiable_path) }
          validate_artifact_path!(verifiable_artifact, potential_conflicts)
        end.empty? && verifiable_artifact[:excludes].empty? && raise(Error::Config, code: :artifact_conflict)
        validate_artifact_path!(verifiable_artifact, artifact[:includes]) if verifiable_artifact[:includes].empty?
      end

      def validate_artifact_path!(verifiable_artifact, potential_conflicts)
        potential_conflicts.all? do |path|
          loop do
            break if verifiable_artifact[:excludes].include?(path) || ((path = File.dirname(path)) == '.')
          end
          verifiable_artifact[:excludes].include?(path)
        end.tap { |res| res || raise(Error::Config, code: :artifact_conflict) }
      end
    end
  end
end
