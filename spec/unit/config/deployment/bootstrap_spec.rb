require_relative '../../../spec_helper'

describe Dapp::Deployment::Config::Directive::Bootstrap do
  include SpecHelper::Common
  include SpecHelper::Config::Deployment

  def dappfile_deployment_bootstrap(&blk)
    dappfile do
      deployment do
        bootstrap do
          instance_eval(&blk)
        end
      end
    end
  end

  def dappfile_app_bootstrap(&blk)
    dappfile do
      deployment do
        app do
          bootstrap do
            instance_eval(&blk)
          end
        end
      end
    end
  end

  def deployment_config
    config
  end

  context 'bootstrap' do
    [:deployment, :app].each do |directive|
      it "#{directive} run" do
        public_send(:"dappfile_#{directive}_bootstrap") do
          run 'cmd'
        end
        expect(public_send(:"#{directive}_config")._deployment._bootstrap._run).to eq ['cmd']
      end

      it "#{directive} dimg" do
        public_send(:"dappfile_#{directive}_bootstrap") do
          dimg 'backend'
        end
        expect(public_send(:"#{directive}_config")._deployment._bootstrap._dimg).to eq 'backend'
      end
    end
  end
end
