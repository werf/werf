require_relative '../spec_helper'

describe Dapp::Dimg::GitRepo do
  include SpecHelper::Common
  include SpecHelper::Dimg
  include SpecHelper::Git

  before :each do
    stub_dimg
  end

  def git_init(git_dir: '.')
    super
    expect(File.exist?(File.join(git_dir, '.git'))).to be_truthy
  end

  def git_change_and_commit(*args, git_dir: '.', **kwargs)
    @commit_counter ||= 0
    super
    @commit_counter += 1

    expect(git_log(git_dir: git_dir).count).to eq @commit_counter
  end

  def dapp_remote_init
    git_init(git_dir: 'remote')

    @remote = Dapp::Dimg::GitRepo::Remote.new(dimg, 'local_remote', url: 'remote/.git')

    expect(File.exist?(@remote.path)).to be_truthy
    expect(@remote.path.to_s[/.*\/([^\/]*\/[^\/]*)/, 1]).to eq 'git_repo_remote/local_remote'
  end

  it 'Remote#init', test_construct: true do
    dapp_remote_init
  end

  it 'Remote#fetch', test_construct: true do
    dapp_remote_init
    git_change_and_commit(git_dir: 'remote')
    @remote.fetch!
    expect(@remote.latest_commit('master')).to eq git_latest_commit(git_dir: 'remote')
  end

  it 'Own', test_construct: true do
    git_init

    own = Dapp::Dimg::GitRepo::Own.new(dimg)
    expect(own.latest_commit).to eq git_latest_commit

    git_change_and_commit
    expect(own.latest_commit).to eq git_latest_commit
  end
end
