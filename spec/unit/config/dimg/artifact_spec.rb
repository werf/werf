require_relative '../../../spec_helper'

describe Dapp::Dimg::Config::Directive::Artifact do
  include SpecHelper::Common
  include SpecHelper::Config::Dimg

  def dappfile_dimg_with_artifact(&blk)
    dappfile do
      dimg_group do
        artifact do
          instance_eval(&blk)
        end

        dimg 'name'
      end
    end
  end

  context 'positive' do
    it 'shell.build_artifact' do
      dappfile_dimg_with_artifact do
        shell do
          build_artifact do
            run 'cmd'
          end
        end
        export '/cwd'
      end

      expect(dimg_config._artifact.first._config._shell._build_artifact_command).to eq ['cmd']
    end

    [:before, :after].each do |order|
      [:setup, :install].each do |stage|
        it "#{order}_#{stage}_artifact" do
          dappfile_dimg_with_artifact do
            export '/cwd' do
              send(order, stage)
            end
          end

          expect(dimg_config.public_send(:"_#{order}_#{stage}_artifact").first).to_not be_nil
        end
      end
    end

    it 'default cwd (to)' do
      dappfile_dimg_with_artifact do
        export do
          before :setup
          to '/a'
        end
      end

      expect(dimg_config._before_setup_artifact.first._cwd).to eq '/a'
    end

    context 'inheritance' do
      def dappfile_dimg_group(&blk)
        dappfile do
          dimg_group do
            instance_eval(&blk)
          end
        end
      end

      it 'independence' do
        dappfile_dimg_group do
          artifact do
            export do
              to '/to_path1'
            end
          end

          artifact do
            export do
              to '/to_path2'
            end
          end

          dimg 'name'
        end

        expect(dimg_config._artifact.size).to eq 2
        dimg_config._artifact.each do |artifact|
          expect(artifact._config._artifact.size).to eq 0
        end
      end

      it 'dimg group' do
        dappfile_dimg_group do
          artifact do
            export do
              to '/to_path1'
            end
          end

          dimg_group do
            artifact do
              export do
                to '/to_path2'
              end
            end

            dimg 'name'
          end
        end

        expect(dimg_config._artifact.size).to eq 2
        first_artifact, second_artifact = dimg_config._artifact
        expect(first_artifact._config._artifact.size).to eq 0
        expect(second_artifact._config._artifact.size).to eq 1
      end
    end
  end

  context 'negative' do
    [:volume, :expose, :env, :label, :cmd, :onbuild, :workdir, :user, :entrypoint].each do |instruction|
      it "unsupported docker.#{instruction} in artifact body" do
        dappfile_dimg_with_artifact do
          docker do
            send(instruction, 'value')
          end

          export do
            to '/to'
          end
        end

        expect_exception_code(:docker_artifact_unsupported_directive) { dimgs_configs }
      end
    end

    it 'double associated (:stage_artifact_double_associate)' do
      dappfile_dimg_with_artifact do
        export '/cwd' do
          before :setup
          after :setup
        end
      end

      expect_exception_code(:stage_artifact_double_associate) { dimgs_configs }
    end

    it 'not supported associated stage (:stage_artifact_not_supported_associated_stage)' do
      dappfile_dimg_with_artifact do
        export '/cwd' do
          before :from
        end
      end

      expect_exception_code(:stage_artifact_not_supported_associated_stage) { dimgs_configs }
    end
  end

  context 'import' do
    context 'positive' do
      it 'same context' do
        dappfile do
          dimg do
            artifact('name') do
              shell do
                build_artifact do
                  run 'cmd'
                end
              end
            end

            import('name') do
              to '/to'
            end
          end
        end

        expect(dimg_config._artifact.first._config._shell._build_artifact_command).to eq ['cmd']
      end

      it 'unrelated context but until using import' do
        dappfile do
          dimg_group do
            dimg_group do
              artifact('name') do
                shell do
                  build_artifact do
                    run 'cmd'
                  end
                end
              end
            end

            dimg_group do
              import('name') do
                to '/to'
              end

              dimg
            end
          end
        end

        expect(dimg_config._artifact.first._config._shell._build_artifact_command).to eq ['cmd']
      end
    end

    context 'negative' do
      it 'import nonexistent artifact (:artifact_not_found)' do
        dappfile do
          dimg do
            import('name')
          end
        end

        expect_exception_code(:artifact_not_found) { dimgs_configs }
      end

      it 'artifact defined after import (:artifact_not_found)' do
        dappfile do
          dimg do
            import('name')
            artifact('name')
          end
        end

        expect_exception_code(:artifact_not_found) { dimgs_configs }
      end

      it 'conflict between artifacts names (:artifact_already_exists)' do
        dappfile do
          dimg do
            artifact('name')
            artifact('name')
          end
        end

        expect_exception_code(:artifact_already_exists) { dimgs_configs }
      end
    end
  end
end
