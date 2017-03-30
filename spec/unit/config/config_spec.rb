require_relative '../../spec_helper'

describe Dapp::Config::Config do
  include SpecHelper::Common
  include SpecHelper::Config

  context 'dev_mode' do
    it 'base (1)' do
      dappfile {}
      expect(config._dev_mode).to eq false
    end

    it 'base (2)' do
      dappfile { dev_mode }
      expect(config._dev_mode).to eq true
    end
  end
end
