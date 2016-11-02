module Dapp
  module Build
    module Stage
      # ArtifactBase
      class ArtifactBase < Base
        def dependencies
          artifacts_signatures
        end

        def prepare_image
          super
          artifacts_dimgs_build!
          artifacts_labels = {}
          artifacts.each do |artifact|
            apply_artifact(artifact, image)
            artifacts_labels["dapp-artifact-#{artifact[:name]}".to_sym] = artifact[:dimg].send(:last_stage).image.built_id
          end
          image.add_service_change_label artifacts_labels
        end

        def images
          [image].concat(artifacts.map { |artifact| artifact[:dimg].images }.flatten)
        end

        protected

        def should_not_be_detailed?
          true
        end

        def ignore_log_commands?
          true
        end

        def artifacts
          @artifacts ||= begin
            dimg.config.public_send("_#{name}").map do |artifact|
              artifact_dimg = Dapp::Artifact.new(config: artifact._config,
                                                 project: dimg.project,
                                                 ignore_git_fetch: dimg.ignore_git_fetch)
              { name: artifact._config._name, options: artifact._artifact_options, dimg: artifact_dimg }
            end
          end
        end

        def artifacts_signatures
          artifacts.map { |artifact| hashsum [artifact[:dimg].signature, artifact[:options]] }
        end

        def artifacts_dimgs_build!
          artifacts.each do |artifact|
            process = dimg.project.t(code: 'process.artifact_building', data: { name: artifact[:name] })
            dimg.project.log_secondary_process(process,
                                               short: !dimg.project.log_verbose?,
                                               quiet: dimg.artifact? && !dimg.project.log_verbose?) do
              dimg.project.with_log_indent do
                artifact[:dimg].build!
              end
            end
          end
        end

        # rubocop:disable Metrics/AbcSize
        def run_artifact_dimg(dimg, artifact_name, commands)
          docker_options = ['--rm',
                            "--volume #{dimg.tmp_path('artifact', artifact_name)}:#{dimg.container_tmp_path(artifact_name)}",
                            "--volumes-from #{dimg.project.base_container}",
                            "--entrypoint #{dimg.project.bash_path}"]
          dimg.project.log_secondary_process(dimg.project.t(code: 'process.artifact_copy',
                                                            data: { name: artifact_name }),
                                             short: true,
                                             quiet: dimg.artifact? && !dimg.project.log_verbose?) do
            dimg.run(docker_options, [%(-ec '#{dimg.project.shellout_pack(commands)}')])
          end
        end
        # rubocop:enable Metrics/AbcSize
      end # ArtifactBase
    end # Stage
  end # Build
end # Dapp
