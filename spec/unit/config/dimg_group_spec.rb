require_relative '../../spec_helper'

describe Dapp::Config::DimgGroup do
  include SpecHelper::Common
  include SpecHelper::Config

  def dappfile_dimg_group(&blk)
    dappfile do
      dimg_group do
        instance_eval(&blk)
      end
    end
  end

  context 'inheritance' do
    context 'shell' do
      [:install, :setup, :before_install, :before_setup].each do |stage|
        it "#{stage}_command" do
          dappfile_dimg_group do
            shell do
              send(stage) do
                command 'cmd1'
              end
            end

            dimg 'name' do
              shell do
                send(stage) do
                  command 'cmd2'
                end
              end
            end
          end

          expect(dimg._shell.public_send("_#{stage}_command")).to eq %w(cmd1 cmd2)
        end

        it "#{stage}_version" do
          dappfile_dimg_group do
            shell do
              send(stage) do
                version 'version1'
              end
            end

            dimg 'name' do
              shell do
                send(stage) do
                  version 'version2'
                end
              end
            end
          end

          expect(dimg._shell.public_send("_#{stage}_version")).to eq 'version2'
        end
      end
    end

    context 'docker' do
      [:volume, :expose, :cmd, :onbuild].each do |attr|
        it attr do
          dappfile_dimg_group do
            docker do
              send(attr, 'value1')
            end

            dimg 'name' do
              docker do
                send(attr, 'value2')
              end
            end
          end

          expect(dimg._docker.send("_#{attr}")).to eq %w(value1 value2)
        end
      end

      [:env, :label].each do |attr|
        it attr do
          dappfile_dimg_group do
            docker do
              send(attr, v1: 1, v2: 1)
            end

            dimg 'name' do
              docker do
                send(attr, v2: 2, v3: 2)
              end
            end
          end

          expect(dimg._docker.send("_#{attr}")).to eq ({ v1: 1, v2: 2, v3: 2 })
        end
      end

      [:workdir, :user].each do |attr|
        it attr do
          dappfile_dimg_group do
            docker do
              send(attr, 'value1')
            end

            dimg 'name' do
              docker do
                send(attr, 'value2')
              end
            end
          end

          expect(dimg._docker.send("_#{attr}")).to eq 'value2'
        end
      end
    end

    it 'mount' do
      dappfile_dimg_group do
        mount '/to' do
          '/from'
        end

        dimg 'name' do
          mount '/to' do
            '/from'
          end
        end
      end

      expect(dimg._mount.size).to eq 2
    end

    it 'artifact' do
      dappfile_dimg_group do
        artifact do
          export do
            to '/to_path2'
          end
        end

        dimg 'name' do
          artifact do
            export do
              to '/to_path2'
            end
          end
        end
      end

      expect(dimg._artifact.size).to eq 2
    end
  end

  context 'warning' do
    def stubbed_project
      super.tap do |instance|
        allow(instance).to receive(:log_config_warning) { |*args, **kwargs| puts kwargs[:desc][:code] }
      end
    end

    it 'artifact after dimg' do
      dappfile do
        dimg_group do
          dimg 'name'
          artifact
        end
      end

      expect { dimg }.to output("wrong_using_directive\n").to_stdout_from_any_process
    end

    [:docker, :shell, :chef].each do |directive|
      it "#{directive} after dimg" do
        dappfile do
          dimg_group do
            dimg 'name'
            send(directive)
          end
        end

        expect { dimg }.to output("wrong_using_base_directive\n").to_stdout_from_any_process
      end
    end
  end
end
