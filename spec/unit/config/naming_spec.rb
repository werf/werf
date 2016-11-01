require_relative '../../spec_helper'

describe Dapp::Config::DimgGroupMain do
  include SpecHelper::Common
  include SpecHelper::Config

  context 'positive' do
    it 'dimg without name (1)' do
      dappfile do
        dimg
      end
      expect(dimg._name).to eq nil
    end

    it 'dimg name' do
      dappfile do
        dimg 'sample'
      end
      expect(dimg._name).to eq 'sample'
    end
  end

  context 'negative' do
    it 'dimg without name (1)' do
      dappfile do
        dimg
        dimg
      end
      expect_exception_code(:dimg_name_required) { dimg }
    end

    it 'dimg without name (2)' do
      dappfile do
        dimg_group do
          dimg
          dimg
        end
      end
      expect_exception_code(:dimg_name_required) { dimg }
    end

    it 'dimg without name (3)' do
      dappfile do
        dimg_group do
          dimg
        end
        dimg_group do
          dimg
        end
      end
      expect_exception_code(:dimg_name_required) { dimg }
    end

    it 'dimg incorrect name' do
      dappfile do
        dimg 'test;'
      end
      expect_exception_code(:dimg_name_incorrect) { dimg }
    end
  end
end
