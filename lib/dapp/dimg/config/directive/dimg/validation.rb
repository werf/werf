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
                raise Error::Config, code: :stage_artifact_not_associated unless _import_artifact.empty?
              end
            end

            def directives_validate!
              passed_directives.each do |v|
                next if (value = instance_variable_get(v)).nil?
                Array(value).each do |elm|
                  elm.validate! if elm.respond_to?(:validate!)
                end
              end
              _mount.map(&:_to).tap do |mounts_points|
                mounts_points.each do |path|
                  raise Error::Config, code: :mount_duplicate_to, data: { path: path } if mounts_points.count(path) > 1
                end
              end
            end

            def validate_scratch_directives!
              directives = [[:_shell, :shell], [:_chef, :chef], [:_git_artifact, :git],
                            [:_tmp_dir_mount, :mount], [:_build_dir_mount, :mount], [:_custom_dir_mount, :mount]]
              directives.each do |name, user_name|
                raise Error::Config,
                      code: :scratch_unsupported_directive,
                      data: { directive: user_name } unless public_send(name).empty?
              end

              docker_directives = [:_expose, :_env, :_cmd, :_onbuild, :_workdir, :_user, :_entrypoint]
              docker_directives.each do |directive|
                value = _docker.public_send(directive)
                raise Error::Config,
                      code: :scratch_unsupported_directive,
                      data: { directive: "docker.#{directive}" } unless value.nil? || value.empty?
              end
            end

            def validate_scratch_artifacts!
              raise Error::Config, code: :scratch_artifact_associated unless _associated_artifacts.empty?
              raise Error::Config, code: :scratch_artifact_required if _import_artifact.empty?
              _import_artifact.each do |artifact|
                raise Error::Config, code: :scratch_artifact_docker_from if artifact._config._docker._from.nil?
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
                artifacts.select { |a| a[:to] == verifiable_artifact[:to] }.each do |artifact|
                  next if verifiable_artifact[:index] == artifact[:index]
                  begin
                    validate_artifact!(verifiable_artifact, artifact)
                    validate_artifact!(artifact, verifiable_artifact)
                  rescue Error::Config => e
                    conflict_between_artifacts!(artifact, verifiable_artifact) if e.net_status[:code] == :artifact_conflict
                    raise
                  end
                end
              end
            end

            def conflict_between_artifacts!(*formatted_artifacts)
              artifacts = formatted_artifacts.flatten.map { |formatted_artifact| formatted_artifact[:related_artifact] }
              dappfile_context = artifacts.map do |artifact|
                artifact_directive = []
                artifact_directive << begin
                  if artifact.is_a? Artifact::Export
                    "artifact.export('#{artifact._cwd}') do"
                  else
                    "git#{"('#{artifact._url}')" if artifact.respond_to?(:_url)}.add('#{artifact._cwd}') do"
                  end
                end
                [:include_paths, :exclude_paths].each do |directive|
                  next if (paths = artifact.send("_#{directive}")).empty?
                  artifact_directive << "  #{directive} '#{paths.join("', '")}'"
                end
                artifact_directive << "  to '#{artifact._to}'"
                artifact_directive << 'end'
                artifact_directive.join("\n")
              end.join("\n\n")
              raise Error::Config, code: :artifact_conflict, data: { dappfile_context: dappfile_context }
            end

            def validate_artifact_format(artifacts)
              artifacts.map do |a|
                path_format = proc { |path| File.expand_path(File.join('/', path, '/'))[1..-1] }
                path_format.call(a._to) =~ %r{^([^\/]*)\/?(.*)$}

                to = Regexp.last_match(1)
                include_exclude_path_format = proc do |path|
                  paths = [].tap do |arr|
                    arr << Regexp.last_match(2) unless Regexp.last_match(2).empty?
                    arr << path
                  end
                  path_format.call(File.join(*paths))
                end

                include_paths = [].tap do |arr|
                  if a._include_paths.empty? && !Regexp.last_match(2).empty?
                    arr << Regexp.last_match(2)
                  else
                    arr.concat(a._include_paths.dup.map(&include_exclude_path_format))
                  end
                end
                exclude_paths = a._exclude_paths.dup.map(&include_exclude_path_format)

                {
                  index: artifacts.index(a),
                  to: to,
                  include_paths: include_paths,
                  exclude_paths: exclude_paths,
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

              raise Error::Config, code: :artifact_conflict if cases.any?
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
