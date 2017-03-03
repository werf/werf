require_relative '../../spec_helper'

describe Dapp::Dimg::Config::Directive::Chef do
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
    dappfile_dimg_chef do
      line("dimod 'dimod-common'")
      line("dimod 'dimod-nginx'")
      line("dimod 'dimod-extra', '~> 0.1.0', git: 'https://github.com/flant/dimod-extra'")
      line("dimod 'dimod-example', path: '../dimod-example'")
    end

    expect(dimg._chef._dimod).to eq(['dimod-common', 'dimod-nginx', 'dimod-extra', 'dimod-example'])

    expect(dimg._chef._cookbook).to eq({
      'dimod-common' => {name: 'dimod-common'},
      'dimod-nginx' => {name: 'dimod-nginx'},
      'dimod-extra' => {name: 'dimod-extra', version_constraint: '~> 0.1.0', git: 'https://github.com/flant/dimod-extra'},
      'dimod-example' => {name: 'dimod-example', path: '../dimod-example'}
    })
  end

  it 'cookbook' do
    dappfile_dimg_chef do
      line("cookbook 'apt'")
      line("cookbook 'ehlo', '~> 0.1.0', git: 'https://github.com/flant/ehlo'")
      line("cookbook 'wrld', path: '../wrld'")
    end

    expect(dimg._chef._cookbook).to eq({
      'apt' => {name: 'apt'},
      'ehlo' => {name: 'ehlo', version_constraint: '~> 0.1.0', git: 'https://github.com/flant/ehlo'},
      'wrld' => {name: 'wrld', path: '../wrld'}
    })
  end

  it 'recipe' do
    dappfile_dimg_chef do
      line("recipe 'main'")
      line("recipe 'hello'")
      line("recipe 'world'")
    end

    expect(dimg._chef._recipe).to eq(['main', 'hello', 'world'])
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
