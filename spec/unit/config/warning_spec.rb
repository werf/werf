require_relative '../../spec_helper'

describe Dapp::Config::DimgGroupMain do
  include SpecHelper::Common
  include SpecHelper::Config

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
