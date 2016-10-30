require_relative '../spec_helper'

describe Dapp::Config::DimgGroupMain do
  include SpecHelper::Common
  include SpecHelper::Config

  def project_name
    File.basename(Dir.getwd)
  end

  def dimg_name(name)
    File.join(project_name, name)
  end

  context 'base' do
    context 'positive' do
      it 'dimg without name (1)' do
        dappfile do
          dimg
        end
        expect(dimg._name).to eq project_name
      end

      it 'dimg without name (2)' do
        dappfile do
          dimg_group do
            dimg
          end
        end
        expect(dimg._name).to eq project_name
      end

      it 'dimg name' do
        dappfile do
          dimg 'sample'
        end
        expect(dimg._name).to eq dimg_name('sample')
      end
    end

    context 'negative' do
      it 'dimg without name' do
        dappfile do
          dimg
          dimg
        end
        expect { dimgs }.to raise_error RuntimeError
      end
    end
  end
end
