module Dapp
  module Dimg
    module Config
      module Directive
        class Dimg < Base
          module Validation
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
                artifacts.select { |a| a[:to] == verifiable_artifact[:to] }.each do |artifact|
                  next if verifiable_artifact[:index] == artifact[:index]
                  validate_artifact!(verifiable_artifact, artifact)
                  validate_artifact!(artifact, verifiable_artifact)
                end
              end
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
                  exclude_paths: exclude_paths
                }
              end
            end

            def validate_artifact!(verifiable_artifact, artifact)
              verifiable_artifact[:include_paths].each do |verifiable_path|
                potential_conflicts = artifact[:include_paths].select { |path| path.start_with?(verifiable_path) }
                validate_artifact_path!(verifiable_artifact, potential_conflicts)
              end

              if verifiable_artifact[:include_paths].empty?
                if artifact[:include_paths].empty? || verifiable_artifact[:exclude_paths].empty?
                  raise ::Dapp::Error::Config, code: :artifact_conflict
                else
                  validate_artifact_path!(verifiable_artifact, artifact[:include_paths])
                end
              end
            end

            def validate_artifact_path!(verifiable_artifact, potential_conflicts)
              raise ::Dapp::Error::Config, code: :artifact_conflict unless begin
                potential_conflicts.all? do |path|
                  loop do
                    break if verifiable_artifact[:exclude_paths].include?(path) || ((path = File.dirname(path)) == '.')
                  end
                  verifiable_artifact[:exclude_paths].include?(path)
                end
              end
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
