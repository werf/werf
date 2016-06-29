require_relative '../spec_helper'

describe Dapp::GitRepo do
  before :each do
    @application = instance_double('Dapp::Application')

    allow(@application).to receive(:build_path) do |*args|
      File.join(*args)
    end

    allow(@application).to receive(:shellout!) do |*args, **kwargs|
      shellout(*args, **kwargs)
    end

    allow(@application).to receive(:filelock).and_yield
  end

  def chronicler_init
    @chrono = Dapp::GitRepo::Chronicler.new(@application, 'chrono')
    expect(File.exist?('chrono')).to be_truthy
    expect(File.exist?('chrono.git')).to be_truthy
  end

  def chronicler_cleanup
    @chrono.cleanup!
    expect(File.exist?('chrono')).to be_falsy
    expect(File.exist?('chrono.git')).to be_falsy
  end

  def chronicler_commit(data)
    @commit_counter ||= 1

    if !File.exist?('chrono/test.txt') || File.read('chrono/test.txt') != data
      @commit_counter += 1
      File.write 'chrono/test.txt', data
    end

    @chrono.commit!
    expect(`git -C chrono.git rev-list --all --count`).to eq "#{@commit_counter}\n"
  end

  it 'Chronicler#create_and_delete', test_construct: true do
    chronicler_init
    chronicler_cleanup
  end

  it 'Chronicler#commit', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')
    chronicler_commit('Some another text')
    chronicler_cleanup
  end

  it 'Chronicler#empty_commit', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')
    chronicler_commit('Some text')
    chronicler_commit('Some another text')
    chronicler_cleanup
  end

  it 'Chronicler#_commit_at_and_latest_commit', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')
    expect(Time.now - @chrono.commit_at(@chrono.latest_commit)).to be < 2
    chronicler_cleanup
  end

  def remote_init(**kwargs)
    chronicler_init
    chronicler_commit('Some text')
    @remote = Dapp::GitRepo::Remote.new(@application, 'remote', url: 'chrono.git', **kwargs)
    expect(File.exist?('remote.git')).to be_truthy
  end

  def remote_cleanup
    @remote.cleanup!
    expect(File.exist?('remote')).to be_falsy
    expect(File.exist?('remote.git')).to be_falsy
  end

  it 'Remote#init', test_construct: true do
    remote_init
    remote_cleanup
  end

  it 'Remote#ssh', test_construct: true do
    shellout 'ssh-keygen -b 1024 -f key -P ""'
    allow(@application).to receive(:home_path).and_return('')
    remote_init ssh_key_path: 'key'
    remote_cleanup
  end

  it 'Remote#fetch', test_construct: true do
    remote_init
    chronicler_commit('Some another text')
    @remote.fetch!
    expect(`git -C remote.git rev-list --all --count`).to eq "#{@commit_counter}\n"
    remote_cleanup
  end

  it 'Own', test_construct: true do
    chronicler_init
    chronicler_commit('Some text')

    allow(@application).to receive(:home_path).and_return('chrono')
    @own = Dapp::GitRepo::Own.new(@application)
    expect(@own.latest_commit).to eq @chrono.latest_commit

    chronicler_commit('Some another text')
    expect(@own.latest_commit).to eq @chrono.latest_commit
  end
end
