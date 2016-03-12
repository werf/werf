require_relative 'spec_helper'

describe Dapp::Atomizer do
  before :each do
    @builder = instance_double('Dapp::Builder')
    allow(@builder).to receive(:register_atomizer)
  end

  def create_atomizer
    @atomizer = Dapp::Atomizer.new @builder, 'atomizer_file'
    expect(@builder).to have_received(:register_atomizer).with(@atomizer)
  end

  def add_file
    FileUtils.touch 'foo'
    @atomizer << 'foo'
    expect(File.readlines('atomizer_file')).to contain_exactly "foo\n"
  end

  def commit_atomizer
    @atomizer.commit!
    expect(File.read('atomizer_file')).to eq ''
    expect(File.exist?('foo')).to be_truthy
  end

  def close_atomizer
    @atomizer.instance_variable_get(:@file).close
    expect(@atomizer.instance_variable_get(:@file).closed?).to be_truthy
  end

  it '#commit_flow', test_construct: true do
    create_atomizer
    add_file

    commit_atomizer
    close_atomizer

    create_atomizer
    expect(File.exist?('foo')).to be_truthy
  end

  it '#rollback_flow', test_construct: true do
    create_atomizer
    add_file

    close_atomizer
    expect(File.readlines('atomizer_file')).to contain_exactly "foo\n"

    create_atomizer
    expect(File.read('atomizer_file')).to eq ''
    expect(File.exist?('foo')).to be_falsy
  end

  it '#rollback_commit_flow', test_construct: true do
    create_atomizer
    add_file

    close_atomizer
    expect(File.readlines('atomizer_file')).to contain_exactly "foo\n"

    create_atomizer
    expect(File.read('atomizer_file')).to eq ''
    expect(File.exist?('foo')).to be_falsy

    add_file
    commit_atomizer

    close_atomizer

    create_atomizer
    expect(File.exist?('foo')).to be_truthy
  end

  it '#locks', test_construct: true do
    Dapp::Atomizer.lock_timeout = 0.01
    Dapp::Atomizer.class_eval do
      def on_error
        throw :exit
      end
    end

    create_atomizer
    expect { create_atomizer }.to throw_symbol(:exit)
  end
end
