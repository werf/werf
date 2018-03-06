module Dapp
  module Dimg
    module Build
      module Stage
        class ArtifactBase < Base
          def dependencies
            @dependencies ||= artifacts_signatures
          end

          def prepare_image
            super do
              artifacts_dimgs_build!
              artifacts_labels = {}
              artifacts.each do |artifact|
                apply_artifact(artifact, image)
                artifacts_labels["dapp-artifact-#{artifact[:dimg].name}".to_sym] = artifact[:dimg].last_stage.image.built_id
              end
              image.add_service_change_label artifacts_labels
            end
          end

          def artifacts
            @artifacts ||= begin
              dimg.config.public_send("_#{name}").map do |artifact|
                artifact_dimg = dimg.dapp.artifact_dimg(config: artifact._config,
                                                        ignore_git_fetch: dimg.ignore_git_fetch)
                { options: artifact._artifact_options, dimg: artifact_dimg }
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
            artifacts.uniq { |artifact| artifact[:dimg] }.each do |artifact|
              process = dimg.dapp.t(code: 'process.artifact_building', data: { name: artifact[:dimg].name })
              dimg.dapp.log_secondary_process(process) { artifact[:dimg].build! }
            end
          end

          def run_artifact_dimg(artifact_dimg, artifact_name, commands)
            docker_options = ['--rm',
                              "--volume #{dimg.tmp_path('artifact', artifact_name)}:#{artifact_dimg.container_tmp_path(artifact_name)}",
                              "--volumes-from #{dimg.dapp.toolchain_container}",
                              "--volumes-from #{dimg.dapp.base_container}",
                              "--entrypoint #{dimg.dapp.bash_bin}"]
            dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.artifact_copy', data: { name: artifact_name }), short: true) do
              artifact_dimg.run(docker_options, [%(-ec '#{dimg.dapp.shellout_pack(commands)}')])
            end
          rescue Error::Dimg => e
            raise unless e.net_status[:code] == :dimg_not_run
            raise Error::Build, code: :export_failed, data: { artifact_name: artifact_name }
          end
          # rubocop:enable Metrics/AbcSize
        end # ArtifactBase
      end # Stage
    end # Build
  end # Dimg
end # Dapp
