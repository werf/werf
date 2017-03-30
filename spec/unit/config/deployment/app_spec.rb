require_relative '../../../spec_helper'

describe Dapp::Deployment::Config::Directive::App do
  include SpecHelper::Common
  include SpecHelper::Config::Deployment

  def dappfile_app(&blk)
    dappfile do
      deployment do
        app do
          instance_eval(&blk)
        end
      end
    end
  end

  context 'app' do
    context 'positive' do
      it 'definition' do
        dappfile do
          deployment do
            app
          end
        end
        expect { config }.to_not raise_error
      end

      it 'name' do
        dappfile do
          deployment do
            app 'name'
          end
        end
        expect(app_config._name).to eq 'name'
      end
    end

    context 'negative' do
      it 'incorrect definition' do
        dappfile do
          app
        end
        expect { config }.to raise_error NoMethodError
      end

      %w(name- Name).each do |incorrect_name|
        it "incorrect name `#{incorrect_name}`" do
          dappfile do
            deployment do
              app incorrect_name
            end
          end
          expect_exception_code(:app_name_incorrect) { config }
        end
      end
    end
  end

  context 'dimg' do
    context 'positive' do
      it 'identical names' do
        dappfile do
          dimg 'backend'

          deployment do
            app do
              dimg 'backend'
            end
          end
        end
        expect { deployment_config_validate! }.to_not raise_error
      end

      it 'identical names (nameless)' do
        dappfile do
          dimg

          deployment do
            app do
              dimg nil
            end
          end
        end
        expect { deployment_config_validate! }.to_not raise_error
      end
    end

    context 'negative' do
      it 'not identical names (:app_dimg_not_found)' do
        dappfile do
          dimg 'backend'

          deployment do
            app do
              dimg nil
            end
          end
        end
        expect_exception_code(:app_dimg_not_found) { deployment_config_validate! }
      end
    end
  end

  it 'namespace' do
    dappfile_app { namespace('name') }
    expect { config }.to_not raise_error
  end

  it 'expose' do
    dappfile_app { expose }
    expect { config }.to_not raise_error
  end

  [:bootstrap, :migrate, :run].each do |sub_directive|
    it sub_directive do
      dappfile_app { send(sub_directive, 'cmd') }
      expect(app_config.send(:"_#{sub_directive}")).to eq 'cmd'
    end
  end
end
