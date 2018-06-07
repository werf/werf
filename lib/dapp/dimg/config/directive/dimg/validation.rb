module Dapp
  module Dimg
    module Config
      module Directive
        class Dimg < Base
          module Validation
            include Helper::Trivia

            def validate!
              directives_validate!
              validate_scratch!
              validate_artifacts!
              validate_artifacts_artifacts!
            end

            protected

            def validate_scratch!
              if _docker._from.nil?
                validate_scratch_directives!
                validate_scratch_artifacts!
              else
                raise ::Dapp::Error::Config, code: :stage_artifact_not_associated unless _import_artifact.empty?
              end
            end

            def directives_validate!
              passed_directives.each do |v|
                next if (value = instance_variable_get(v)).nil?
                Array(value).each do |elm|
                  elm.validate! if elm.respond_to?(:validate!)
                end
              end
              _context_artifact_groups.each(&:validate!)
              _mount.map(&:_to).tap do |mounts_points|
                mounts_points.each do |path|
                  raise ::Dapp::Error::Config, code: :mount_duplicate_to, data: { path: path } if mounts_points.count(path) > 1
                end
              end
            end

            def validate_scratch_directives!
              directives = [[:_shell, :shell], [:_chef, :chef], [:_git_artifact, :git],
                            [:_tmp_dir_mount, :mount], [:_build_dir_mount, :mount], [:_custom_dir_mount, :mount]]
              directives.each do |name, user_name|
                raise ::Dapp::Error::Config, code: :scratch_unsupported_directive,
                                             data: { directive: user_name } unless public_send(name).empty?
              end

              docker_directives = [:_expose, :_env, :_cmd, :_onbuild, :_workdir, :_user, :_entrypoint]
              docker_directives.each do |directive|
                value = _docker.public_send(directive)
                raise ::Dapp::Error::Config, code: :scratch_unsupported_directive,
                                             data: { directive: "docker.#{directive}" } unless value.nil? || value.empty?
              end
            end

            def validate_scratch_artifacts!
              raise ::Dapp::Error::Config, code: :scratch_artifact_associated unless _associated_artifacts.empty?
              raise ::Dapp::Error::Config, code: :scratch_artifact_required if _import_artifact.empty?
              _import_artifact.each do |artifact|
                raise ::Dapp::Error::Config, code: :scratch_artifact_docker_from if artifact._config._docker._from.nil?
              end
            end

            def validate_artifacts_artifacts!
              _artifact.each { |artifact_dimg| artifact_dimg._config.validate! }
            end

            def validate_artifacts!
              artifacts = validate_artifact_format(all_artifacts)
              loop do
                break if artifacts.empty?
                verifiable_artifact = artifacts.shift
                artifacts.each do |artifact|
                  next if verifiable_artifact[:index] == artifact[:index]
                  begin
                    validate_artifact!(verifiable_artifact, artifact)
                    validate_artifact!(artifact, verifiable_artifact)
                  rescue ::Dapp::Error::Config => e
                    conflict_between_artifacts!(artifact, verifiable_artifact) if e.net_status[:code] == :artifact_conflict
                    raise
                  end
                end
              end
            end

            def conflict_between_artifacts!(*formatted_artifacts)
              artifacts = formatted_artifacts.flatten.map { |formatted_artifact| formatted_artifact[:related_artifact] }
              artifact_imports = []
              git_artifacts    = []
              artifacts.map do |artifact|
                (artifact.is_a?(Artifact::Export) ? artifact_imports : git_artifacts) << artifact
              end

              common_artifact_dsl = proc do |a|
                [].tap do |context|
                  context << "  to '#{a._to}'"
                  [:include_paths, :exclude_paths].each do |directive|
                    next if (paths = a.send("_#{directive}")).empty?
                    context << "  #{directive} '#{paths.join("', '")}'"
                  end
                end
              end

              import_dsl = proc do |ga|
                [].tap do |context|
                  context << "artifact.export('#{ga._cwd}') do"
                  context.concat(common_artifact_dsl.call(ga))
                  context << 'end'
                end
              end

              git_artifact_dsl = proc do |ga|
                [].tap do |context|
                  context << "git#{"('#{ga._url}')" if ga.respond_to?(:_url)}.add('#{ga._cwd}') do"
                  context.concat(common_artifact_dsl.call(ga))
                  context << 'end'
                end
              end

              common_artifact_yaml = proc do |a|
                [].tap do |context|
                  context << "  to: #{a._to}"
                  {
                    include_paths: 'includePaths',
                    exclude_paths: 'excludePaths',
                  }.each do |directive, yamlDirective|
                    next if (paths = a.send("_#{directive}")).empty?
                    context << "  #{yamlDirective}: [#{paths.join(", ")}]"
                  end
                end
              end

              import_yaml = proc do |a|
                [].tap do |context|
                  context << "- artifact: #{a._config._name}"
                  context << "  add: #{a._cwd}"
                  context.concat(common_artifact_yaml.call(a))
                end
              end

              git_artifact_yaml = proc do |ga|
                [].tap do |context|
                  if ga.respond_to?(:_url)
                    context << "- url: #{ga._url}"
                    context << "  add: #{ga._cwd}"
                  else
                    context << "- add: #{ga._cwd}"
                  end
                  context.concat(common_artifact_yaml.call(ga))
                end
              end

              dappfile_context = [].tap do |msg|
                if artifact_imports.count != 0
                  if dapp
                    artifact_imports.each { |a| msg.concat(import_dsl.call(a)) }
                  else
                    msg << 'import:'
                    artifact_imports.each { |a| msg.concat(import_yaml.call(a)) }
                  end
                end

                if git_artifacts.count != 0
                  if dapp
                    git_artifacts.each { |a| msg.concat(git_artifact_dsl.call(a)) }
                  else
                    msg << 'git:'
                    git_artifacts.each { |a| msg.concat(git_artifact_yaml.call(a)) }
                  end
                end
              end.join("\n")
              raise ::Dapp::Error::Config, code: :artifact_conflict, data: { dappfile_context: dappfile_context }
            end

            def validate_artifact_format(artifacts)
              artifacts.map do |a|
                include_paths = begin
                  if a._include_paths.empty?
                    [a._to]
                  else
                    a._include_paths.dup.map { |p| File.join(a._to, p) }
                  end
                end

                {
                  index: artifacts.index(a),
                  include_paths: include_paths,
                  exclude_paths: a._exclude_paths.dup.map { |p| File.join(a._to, p) },
                  related_artifact: a
                }
              end
            end

            def validate_artifact!(verifiable_artifact, artifact)
              cases = []
              cases << verifiable_artifact[:include_paths].any? do |verifiable_path|
                !ignore_path?(verifiable_path, paths: artifact[:include_paths], exclude_paths: artifact[:exclude_paths])
              end
              cases << (verifiable_artifact[:include_paths].empty? && artifact[:include_paths].empty?)

              raise ::Dapp::Error::Config, code: :artifact_conflict if cases.any?
            end

            def _associated_artifacts
              _artifact.select do |art|
                !(art._before.nil? && art._after.nil?)
              end
            end
          end # Validation
          # rubocop:enable Metrics/ModuleLength
        end # Dimg
      end # Directive
    end # Config
  end # Dimg
end # Dapp
