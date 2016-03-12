require_relative 'spec_helper'

describe Dapp::GitRepo do
  before :each do
    @builder = instance_double('Dapp::Builder')

    allow(@builder).to receive(:build_path) do |*args|
      File.join(*args)
    end

    allow(@builder).to receive(:shellout) do |*args, **kwargs|
      Mixlib::ShellOut.new(*args, timeout: 3600, **kwargs).run_command.tap(&:error!)
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

    repo.cleanup!
    expect(File.exist?('test')).to be_falsy
    expect(File.exist?('test.git')).to be_falsy
  end
end
