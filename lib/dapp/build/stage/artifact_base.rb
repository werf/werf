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
          artifacts_applications_build!
          artifacts_labels = {}
          artifacts.each do |artifact|
            apply_artifact(artifact, image)
            artifacts_labels["dapp-artifact-#{artifact[:name]}".to_sym] = artifact[:app].send(:last_stage).image.built_id
          end
          image.add_service_change_label artifacts_labels
        end

        def images
          [image].concat(artifacts.map { |artifact| artifact[:app].images }.flatten)
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
            application.config.public_send("_#{name}").map do |artifact|
              app = Dapp::Artifact.new(config: artifact._config,
                                       project: application.project,
                                       ignore_git_fetch: application.ignore_git_fetch)
              { name: artifact._config._name, options: artifact._artifact_options, app: app }
            end
          end
        end

        def artifacts_signatures
          artifacts.map { |artifact| hashsum [artifact[:app].signature, artifact[:options]] }
        end

        def artifacts_applications_build!
          artifacts.each do |artifact|
            process = application.project.t(code: 'process.artifact_building', data: { name: artifact[:name] })
            application.project.log_secondary_process(process,
                                                      short: !application.project.log_verbose?,
                                                      quiet: application.artifact? && !application.project.log_verbose?) do
              application.project.with_log_indent do
                artifact[:app].build!
              end
            end
          end
        end

        # rubocop:disable Metrics/AbcSize
        def run_artifact_app(app, artifact_name, commands)
          docker_options = ['--rm',
                            "--volume #{application.tmp_path('artifact', artifact_name)}:#{app.container_tmp_path(artifact_name)}",
                            "--volumes-from #{application.project.base_container}",
                            "--entrypoint #{application.project.bash_path}"]
          application.project.log_secondary_process(application.project.t(code: 'process.artifact_copy',
                                                                          data: { name: artifact_name }),
                                                    short: true,
                                                    quiet: application.artifact? && !application.project.log_verbose?) do
            app.run(docker_options, [%(-ec '#{application.project.shellout_pack(commands)}')])
          end
        end
        # rubocop:enable Metrics/AbcSize
      end # ArtifactBase
    end # Stage
  end # Build
end # Dapp
