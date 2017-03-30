require_relative '../../../spec_helper'

describe Dapp::Deployment::Config::Directive::Namespace do
  include SpecHelper::Common
  include SpecHelper::Config::Deployment

  def dappfile_app_namespace(name = 'default', &blk)
    dappfile do
      deployment do
        app do
          namespace(name) do
            instance_eval(&blk) if block_given?
          end
        end
      end
    end
  end

  it 'name' do
    dappfile_app_namespace('ru')
    expect(app_config._namespace.length).to eq 1
    expect(app_config._namespace['ru']).to_not be_nil
  end

  %w(name- Name).each do |incorrect_name|
    it "incorrect name `#{incorrect_name}`" do
      dappfile_app_namespace(incorrect_name)
      expect_exception_code(:namespace_name_incorrect) { config }
    end
  end

  [:environment, :secret_environment].each do |sub_directive|
    it sub_directive do
      dappfile_app_namespace do
        send(sub_directive, A: 0, B: 0)
      end
      expect(app_config._namespace['default'].send(:"_#{sub_directive}")).to eq(A: 0, B: 0)
    end

    it "overriding #{sub_directive}" do
      dappfile_app_namespace do
        send(sub_directive, A: 0)
        send(sub_directive, A: 1)
      end
      expect(app_config._namespace['default'].send(:"_#{sub_directive}")).to eq(A: 1)
    end
  end

  context 'scale' do
    it 'base' do
      value = 1000
      dappfile_app_namespace do
        scale value
      end
      expect(app_config._namespace['default']._scale).to eq(value)
    end

    [-1, 0, 'value'].each do |value|
      it "incorrect value `#{value}` (:unsupported_scale_value)" do
        dappfile_app_namespace do
          scale value
        end
        expect_exception_code(:unsupported_scale_value) { deployment_config_validate! }
      end
    end
  end
end
