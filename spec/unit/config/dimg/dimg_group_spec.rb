require_relative '../../../spec_helper'

describe Dapp::Dimg::Config::Directive::DimgGroup do
  include SpecHelper::Common
  include SpecHelper::Config::Dimg

  def dappfile_dimg_group(&blk)
    dappfile do
      dimg_group do
        instance_eval(&blk)
      end
    end
  end

  context 'inheritance' do
    context 'chef' do
      [:dimod, :recipe].each do |directive|
        it directive do
          dappfile_dimg_group do
            chef do
              send(directive, 'value1')
            end

            dimg 'name' do
              chef do
                send(directive, 'value2')
              end
            end
          end

          expect(dimg_config._chef.public_send("_#{directive}")).to eq %w(value1 value2)
        end
      end

      it 'attributes' do
        dappfile_dimg_group do
          chef do
            line("attributes['k1']['k2'] = 'k1k2value'")
            line("attributes['k1']['k3'] = 'k1k3value'")
          end

          dimg 'name' do
            chef do
              line("attributes['k1']['k2'] = 'k1k2newvalue'")
            end
          end
        end

        expect(dimg_config._chef._attributes).to eq('k1' => { 'k2' => 'k1k2newvalue', 'k3' => 'k1k3value' })
      end
    end

    context 'shell' do
      [:install, :setup, :before_install, :before_setup].each do |stage|
        it "#{stage}_command" do
          dappfile_dimg_group do
            shell do
              send(stage) do
                run 'cmd1'
              end
            end

            dimg 'name' do
              shell do
                send(stage) do
                  run 'cmd2'
                end
              end
            end
          end

          expect(dimg_config._shell.public_send("_#{stage}_command")).to eq %w(cmd1 cmd2)
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

          expect(dimg_config._shell.public_send("_#{stage}_version")).to eq 'version2'
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

          expect(dimg_config._docker.send("_#{attr}")).to eq %w(value1 value2)
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

          expect(dimg_config._docker.send("_#{attr}")).to eq(v1: 1, v2: 2, v3: 2)
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

          expect(dimg_config._docker.send("_#{attr}")).to eq 'value2'
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

      expect(dimg_config._mount.size).to eq 2
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

      expect(dimg_config._artifact.size).to eq 2
    end
  end

  context 'warning' do
    def stubbed_dapp
      super.tap do |instance|
        allow(instance).to receive(:log_config_warning) { |*_args, **kwargs| puts kwargs[:desc][:code] }
      end
    end

    it 'artifact after dimg' do
      dappfile do
        dimg_group do
          dimg 'name'
          artifact
        end
      end

      expect { dimg_config }.to output("wrong_using_directive\n").to_stdout_from_any_process
    end

    [:docker, :shell, :chef].each do |directive|
      it "#{directive} after dimg" do
        dappfile do
          dimg_group do
            dimg 'name'
            send(directive)
          end
        end

        expect { dimg_config }.to output("wrong_using_base_directive\n").to_stdout_from_any_process
      end
    end
  end
end
