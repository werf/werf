require_relative 'spec_helper'

describe Dapp::Atomizer do
  before :each do
    @builder = instance_double('Dapp::Builder')
    allow(@builder).to receive(:register_atomizer)
  end

  def create_atomizer
    @atomizer = Dapp::Atomizer.new(@builder, 'atomizer_file')
    expect(@builder).to have_received(:register_atomizer).with(@atomizer)
  end

  def add_file(construct)
    @test_file_path = construct.file 'foo'
    @atomizer << @test_file_path
    expect(File.readlines('atomizer_file')).to contain_exactly "#{@test_file_path}\n"
  end

  def commit_atomizer
    @atomizer.commit!
    expect(File.read('atomizer_file')).to eq ''
    expect(@test_file_path.exist?).to be_truthy
  end

  def close_atomizer
    @atomizer.instance_variable_get(:@file).close
    expect(@atomizer.instance_variable_get(:@file).closed?).to be_truthy
  end

  it '#commit_flow', test_construct: true do |example|
    create_atomizer
    add_file(example.metadata[:construct])

    commit_atomizer
    close_atomizer

    create_atomizer
    expect(@test_file_path.exist?).to be_truthy
  end

  it '#rollback_flow', test_construct: true do |example|
    create_atomizer
    add_file(example.metadata[:construct])

    close_atomizer
    expect(File.readlines('atomizer_file')).to contain_exactly "#{@test_file_path}\n"

    create_atomizer
    expect(File.read('atomizer_file')).to eq ''
    expect(@test_file_path.exist?).to be_falsy
  end

  it '#rollback_commit_flow', test_construct: true do |example|
    create_atomizer
    add_file(example.metadata[:construct])

    close_atomizer
    expect(File.readlines('atomizer_file')).to contain_exactly "#{@test_file_path}\n"

    create_atomizer
    expect(File.read('atomizer_file')).to eq ''
    expect(@test_file_path.exist?).to be_falsy

    add_file(example.metadata[:construct])
    commit_atomizer

    close_atomizer

    create_atomizer
    expect(@test_file_path.exist?).to be_truthy
  end
end
