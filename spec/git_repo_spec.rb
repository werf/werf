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

  def chronicler_init
    @repo = Dapp::GitRepo::Chronicler.new(@builder, 'chrono')
    expect(File.exist?('chrono')).to be_truthy
    expect(File.exist?('chrono.git')).to be_truthy
  end

  def chronicler_cleanup
    @repo.cleanup!
    expect(File.exist?('chrono')).to be_falsy
    expect(File.exist?('chrono.git')).to be_falsy
  end

  def chronicler_commit(data)
    @commit_counter ||= 1

    if File.exist?('chrono/test.txt') && File.read('chrono/test.txt') != data
      @commit_counter += 1
      example.metadata[:construct].file 'chrono/test.txt', data
    end

    @repo.commit!
    expect(`cd chrono; git rev-list --all --count`).to eq "#{@commit_counter}\n"
  end

  it 'Chronicler # create and delete', test_construct: true do
    chronicler_init
    chronicler_cleanup
  end

  it 'Chronicler #commit', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')
    chronicler_commit('Some another text')
    chronicler_cleanup
  end

  it 'Chronicler # empty commit', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')
    chronicler_commit('Some text')
    chronicler_commit('Some another text')
    chronicler_cleanup
  end

  it 'Chronicler # commit_at and latest_commit', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')
    expect(Time.now - @repo.commit_at(@repo.latest_commit)).to be < 2
    chronicler_cleanup
  end
end
