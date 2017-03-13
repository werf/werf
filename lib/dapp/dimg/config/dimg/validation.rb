module Dapp
  module Dimg
    module Config
      class Dimg < Base
        module Validation
          protected

          def validate!
            directives_validate!
            validate_scratch!
            validate_artifacts!
            validate_artifacts_artifacts!
          end

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
                elm.send(:validate!) if elm.methods.include?(:validate!)
              end
            end
          end

          def validate_scratch_directives!
            directives = [[:_shell, :shell], [:_chef, :chef], [:_git_artifact, :git],
                          [:_tmp_dir_mount, :mount], [:_build_dir_mount, :mount]]
            directives.each do |name, user_name|
              raise Error::Config,
                    code: :scratch_unsupported_directive,
                    data: { directive: user_name } unless public_send(name).send(:empty?)
            end

            docker_directives = [:_expose, :_env, :_cmd, :_onbuild, :_workdir, :_user, :_entrypoint]
            docker_directives.each do |directive|
              value = _docker.public_send(directive)
              raise Error::Config,
                    code: :scratch_unsupported_directive,
                    data: { directive: "docker.#{directive}" } unless value.nil? || value.send(:empty?)
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
            artifacts = validate_artifact_format(validated_artifacts)
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
              include_paths = a._include_paths.dup
              include_paths << Regexp.last_match(2) unless Regexp.last_match(2).empty?
              exclude_paths = a._exclude_paths.dup

              {
                index: artifacts.index(a),
                to: to,
                include_paths: include_paths.map(&path_format),
                exclude_paths: exclude_paths.map(&path_format)
              }
            end
          end

          def validate_artifact!(verifiable_artifact, artifact)
            verifiable_artifact[:include_paths].each do |verifiable_path|
              potential_conflicts = artifact[:include_paths].select { |path| path.start_with?(verifiable_path) }
              validate_artifact_path!(verifiable_artifact, potential_conflicts)
            end.empty? && verifiable_artifact[:exclude_paths].empty? && raise(Error::Config, code: :artifact_conflict)
            validate_artifact_path!(verifiable_artifact, artifact[:include_paths]) if verifiable_artifact[:include_paths].empty?
          end

          def validate_artifact_path!(verifiable_artifact, potential_conflicts)
            potential_conflicts.all? do |path|
              loop do
                break if verifiable_artifact[:exclude_paths].include?(path) || ((path = File.dirname(path)) == '.')
              end
              verifiable_artifact[:exclude_paths].include?(path)
            end.tap { |res| res || raise(Error::Config, code: :artifact_conflict) }
          end

          def _associated_artifacts
            _artifact.select do |art|
              !(art._before.nil? && art._after.nil?)
            end
          end

          def validated_artifacts
            _artifact + _git_artifact._local + _git_artifact._remote
          end
        end # Validation
        # rubocop:enable Metrics/ModuleLength
      end # Dimg
    end # Config
  end # Dimg
end # Dapp
