require_relative '../../../spec_helper'

describe Dapp::Deployment::Config::Directive::Expose do
  include SpecHelper::Common
  include SpecHelper::Config::Deployment

  def dappfile_app_expose(&blk)
    dappfile do
      deployment do
        app do
          expose do
            instance_eval(&blk)
          end
        end
      end
    end
  end

  it 'cluster_ip' do
    dappfile_app_expose do
      cluster_ip
    end
    expect(app_config._expose._cluster_ip).to be_truthy
  end

  context 'port' do
    def dappfile_app_expose_port(*args, &blk)
      dappfile_app_expose do
        port(*args) do
          instance_eval(&blk) if block_given?
        end
      end
    end

    context 'positive' do
      it 'definition' do
        values = [80, 8080]
        dappfile_app_expose_port(*values)
        expect(app_config._expose._port.length).to be 1
        expect(app_config._expose._port.first._list).to eq values
      end

      it 'default protocol' do
        dappfile_app_expose_port(80)
        expect(app_config._expose._port.first._protocol).to eq 'TCP'
      end

      it 'udp protocol' do
        dappfile_app_expose_port(80) { udp }
        expect(app_config._expose._port.first._protocol).to eq 'UDP'
      end

      it 'tcp protocol' do
        dappfile_app_expose_port(80) { tcp }
        expect(app_config._expose._port.first._protocol).to eq 'TCP'
      end
    end

    context 'negative' do
      %w(-1 65537).each do |incorrect_value|
        it "unsupported port value `#{incorrect_value}` (:unsupported_port_number)" do
          dappfile_app_expose_port(incorrect_value)
          expect_exception_code(:unsupported_port_number) { config }
        end
      end
    end
  end
end
