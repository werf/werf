module Dapp
  module Dimg
    module Build
      module Stage
        class ArtifactBase < Base
          def dependencies
            @dependencies ||= artifacts_signatures
          end

          def prepare_image
            super
            artifacts_dimgs_build!
            artifacts_labels = {}
            artifacts.each do |artifact|
              apply_artifact(artifact, image)
              artifacts_labels["dapp-artifact-#{artifact[:name]}".to_sym] = artifact[:dimg].last_stage.image.built_id
            end
            image.add_service_change_label artifacts_labels
          end

          def artifacts
            @artifacts ||= begin
              dimg.config.public_send("_#{name}").map do |artifact|
                artifact_dimg = ::Dapp::Dimg::Artifact.new(config: artifact._config,
                                                           dapp: dimg.dapp,
                                                           ignore_git_fetch: dimg.ignore_git_fetch)
                { name: artifact._config._name, options: artifact._artifact_options, dimg: artifact_dimg }
              end
            end
          end

          def artifact?
            true
          end

          protected

          def should_not_be_detailed?
            true
          end

          def ignore_log_commands?
            true
          end

          def artifacts_signatures
            artifacts.map { |artifact| hashsum [artifact[:dimg].signature, artifact[:options]] }
          end

          def artifacts_dimgs_build!
            artifacts.each do |artifact|
              process = dimg.dapp.t(code: 'process.artifact_building', data: { name: artifact[:name] })
              dimg.dapp.log_secondary_process(process) { artifact[:dimg].build! }
            end
          end

          def run_artifact_dimg(artifact_dimg, artifact_name, commands)
            dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.artifact_copy', data: { name: artifact_name }), short: true) do
              image_run_options = {}.tap do |options|
                options[:image]      = artifact_dimg.last_stage.image.built_id
                options[:entrypoint] = dimg.dapp.bash_bin
                options[:cmd]        = ['-ec', dimg.dapp.shellout_pack(commands)]
                options[:hostconfig] = {}.tap do |hostconfig|
                  hostconfig[:mounts] = [{ source: dimg.tmp_path('artifact', artifact_name).tap(&:mkpath), target: artifact_dimg.container_tmp_path(artifact_name), type: :bind }]
                  hostconfig[:volumesfrom] = [dimg.dapp.base_container]
                end
              end
              dimg.dapp.docker_client.container_run(rm: true, **image_run_options)
            end
          end
          # rubocop:enable Metrics/AbcSize
        end # ArtifactBase
      end # Stage
    end # Build
  end # Dimg
end # Dapp
