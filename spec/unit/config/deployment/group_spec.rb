require_relative '../../../spec_helper'

describe Dapp::Deployment::Config::Directive::Group do
  include SpecHelper::Common
  include SpecHelper::Config::Deployment

  def dappfile_deployment(&blk)
    dappfile do
      deployment do
        instance_eval(&blk) if block_given?
      end
    end
  end

  def dappfile_deployment_group(&blk)
    dappfile_deployment do
      group do
        instance_eval(&blk) if block_given?
      end
    end
  end

  context 'base' do
    context 'positive' do
      it 'top level definition' do
        dappfile { deployment }
        expect { config }.to_not raise_error
      end

      it 'definition' do
        dappfile { deployment { group } }
        expect { config }.to_not raise_error
      end
    end

    context 'negative' do
      it 'incorrect top level definition' do
        dappfile { group }
        expect { config }.to raise_error NoMethodError
      end

      it 'incorrect definition' do
        dappfile { deployment { deployment } }
        expect { config }.to raise_error NoMethodError
      end
    end

    context 'directive' do
      it 'app' do
        dappfile_deployment_group { app }
        expect { config }.to_not raise_error
      end

      it 'namespace' do
        dappfile_deployment_group { namespace('name') }
        expect { config }.to_not raise_error
      end

      it 'expose' do
        dappfile_deployment_group { expose }
        expect { config }.to_not raise_error
      end

      [:bootstrap, :migrate, :run].each do |sub_directive|
        it sub_directive do
          dappfile_deployment_group { send(sub_directive, 'cmd') }
          expect { config }.to_not raise_error
        end
      end
    end
  end

  context 'inheritance' do
    context 'namespace' do
      [:environment, :secret_environment].each do |sub_directive|
        it sub_directive do
          dappfile_deployment_group do
            namespace('default') do
              send(sub_directive, A: 0, B: 1)
            end

            app do
              namespace('default') do
                send(sub_directive, A: 2, C: 3)
              end
            end
          end
          expect(app_config._namespace['default'].send(:"_#{sub_directive}")).to eq(A: 2, B: 1, C: 3)
        end

        it 'scale' do
          dappfile_deployment_group do
            namespace('default') do
              scale 1
            end
            app
          end
          expect(app_config._namespace['default']._scale).to eq(1)
        end

        it 'overriding scale' do
          dappfile_deployment_group do
            namespace('default') do
              scale 1
            end

            app do
              namespace('default') do
                scale 2
              end
            end
          end
          expect(app_config._namespace['default']._scale).to eq(2)
        end
      end
    end

    context 'expose' do
      it 'base (1)' do
        dappfile_deployment_group do
          expose do
            port(80) do
              udp
            end
          end
          app
        end
        expect(app_config._expose._cluster_ip).to be_falsey
        expect(app_config._expose._port.first._list).to eq([80])
        expect(app_config._expose._port.first._protocol).to eq(:UDP)
      end

      it 'base (2)' do
        dappfile_deployment_group do
          expose do
            port(80) do
              udp
            end
          end
          app do
            expose do
              cluster_ip
              port(8080)
            end
          end
        end
        expect(app_config._expose._cluster_ip).to be_truthy
        expect(app_config._expose._port.length).to eq(2)
      end
    end

    [:bootstrap, :migrate, :run].each do |sub_directive|
      it sub_directive do
        dappfile_deployment_group do
          send(sub_directive, 'cmd')
          app
        end
        expect(app_config.send(:"_#{sub_directive}")).to eq 'cmd'
      end

      it "overriding #{sub_directive}" do
        dappfile_deployment_group do
          send(sub_directive, 'cmd1')
          app do
            send(sub_directive, 'cmd2')
          end
        end
        expect(app_config.send(:"_#{sub_directive}")).to eq 'cmd2'
      end
    end
  end
end
