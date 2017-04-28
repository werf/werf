require_relative '../../../spec_helper'

describe Dapp::Deployment::Config::Directive::Job do
  include SpecHelper::Common
  include SpecHelper::Config::Deployment

  [:bootstrap, :before_apply_job].each do |job|
    define_method :"dappfile_deployment_#{job}" do |&blk|
      dappfile do
        deployment do
          public_send(job) do
            instance_eval(&blk)
          end
        end
      end
    end

    define_method :"dappfile_app_#{job}" do |&blk|
      dappfile do
        deployment do
          app do
            public_send(job) do
              instance_eval(&blk)
            end
          end
        end
      end
    end
  end

  def deployment_config
    config._deployment
  end

  [:bootstrap, :before_apply_job].each do |job|
    context job do
      [:deployment, :app].each do |directive|
        it "#{directive} run" do
          public_send(:"dappfile_#{directive}_#{job}") do
            run 'cmd'
          end
          expect(public_send(:"#{directive}_config").public_send(:"_#{job}")._run).to eq ['cmd']
        end

        it "#{directive} dimg" do
          public_send(:"dappfile_#{directive}_#{job}") do
            dimg 'backend'
          end
          expect(public_send(:"#{directive}_config").public_send(:"_#{job}")._dimg).to eq 'backend'
        end
      end
    end
  end
end
