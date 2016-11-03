require_relative '../../spec_helper'

describe Dapp::Config::Directive::Chef do
  include SpecHelper::Common
  include SpecHelper::Config

  def dappfile_dimg_chef(&blk)
    dappfile do
      dimg do
        chef do
          instance_eval(&blk) if block_given?
        end
      end
    end
  end

  it 'dimod' do
    expect_array_attribute(:dimod, method(:dappfile_dimg_chef)) do |*args|
      expect(dimg._chef._dimod).to eq args
    end
  end

  it 'recipe' do
    expect_array_attribute(:recipe, method(:dappfile_dimg_chef)) do |*args|
      expect(dimg._chef._recipe).to eq args
    end
  end

  it 'attributes' do
    dappfile_dimg_chef do
      line("attributes['k1']['k2'] = 'k1k2value'")
      line("attributes['k1']['k3'] = 'k1k3value'")
    end

    expect(dimg._chef._attributes).to eq('k1' => { 'k2' => 'k1k2value', 'k3' => 'k1k3value' })
  end

  [:before_install, :install, :before_setup, :setup, :build_artifact].map do |key|
    it "#{key}_attributes" do
      dappfile_dimg_chef do
        line("attributes['k1']['#{key}'] = 'k1#{key}value'")
      end

      expect(dimg._chef.send("__#{key}_attributes")).to eq('k1' => { key.to_s => "k1#{key}value" })
    end
  end
end
