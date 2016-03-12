require_relative 'spec_helper'

describe Dapp::GitRepo do
  before :all do
    shellout 'git config -l | grep "user.email" || git config --global user.email "dapp@flant.com"'
    shellout 'git config -l | grep "user.name" || git config --global user.name "Dapp Dapp"'
  end

  before :each do
    @builder = instance_double('Dapp::Builder')

    allow(@builder).to receive(:build_path) do |*args|
      File.join(*args)
    end

    allow(@builder).to receive(:shellout) do |*args, **kwargs|
      shellout(*args, **kwargs)
    end

    allow(@builder).to receive(:filelock).and_yield
  end

  it 'Chronicler', test_construct: true do |example|
    repo = Dapp::GitRepo::Chronicler.new(@builder, 'test')
    expect(File.exist?('test')).to be_truthy
    expect(File.exist?('test.git')).to be_truthy

    example.metadata[:construct].file 'test/test.txt', 'Some text'
    repo.commit!
    expect(`cd test.git; git rev-list --all --count`).to eq "2\n"

    expect(Time.now - repo.commit_at(repo.latest_commit)).to be < 2

    repo.cleanup!
    expect(File.exist?('test')).to be_falsy
    expect(File.exist?('test.git')).to be_falsy
  end
end
