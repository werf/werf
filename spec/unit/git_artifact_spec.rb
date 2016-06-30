require_relative '../spec_helper'

describe Dapp::GitArtifact do
  include SpecHelpers::Common
  include SpecHelpers::Application
  include SpecHelpers::Git
  include SpecHelpers::GitArtifact

  before :each do
    stub_application
  end

  def tar_files_owners(archive)
    shellout("tar -tvf #{archive}").stdout.lines.map { |s| s.strip.sub(%r(.{11}([^\/]+)\/.*), '\1') }.uniq
  end

  def tar_files_groups(archive)
    shellout("tar -tvf #{archive}").stdout.lines.map { |s| s.strip.sub(%r(.{11}[^\/]+\/([^\s]+).*), '\1') }.uniq
  end

  def artifact_do_test(where_to_add, latest_patch: true, layers: 3, **kwargs)
    git_artifact_init where_to_add, **kwargs
    artifact_archive
    layers.times do |i|
      artifact_layer_patch i + 1
    end
    artifact_latest_patch if latest_patch
  end

  def artifact_expect_clean(id: nil)
    expect(Dir.glob(artifact_filename('{.,_}*', id: id)))
      .to match_array([artifact_filename('.paramshash', id: id), artifact_filename('.atomizer', id: id)])
  end

  it '#archive_only', test_construct: true do
    artifact_do_test '/dest', latest_patch: false, layers: 0
  end

  it '#latest_patch', test_construct: true do
    artifact_do_test '/dest', layers: 0
  end

  it '#layer_patch', test_construct: true do
    artifact_do_test '/dest', latest_patch: false, layers: 1
  end

  it '#layer_patch_and_latest_patch', test_construct: true do
    artifact_do_test '/dest', layers: 1
  end

  it '#multiple_layer_patches_and_latest_patch', test_construct: true do
    artifact_do_test '/dest'
  end

  it '#multiple_artifacts', test_construct: true do
    git_artifact_init '/dest', name: 'a', id: :a
    git_artifact_init '/dest_2', name: 'b', id: :b
    artifact_archive id: :b
    artifact_archive id: :a

    artifact_layer_patch 1, id: :a
    artifact_layer_patch 1, id: :b
    artifact_layer_patch 2, id: :b
    artifact_layer_patch 2, id: :a

    artifact_latest_patch id: :b
    artifact_latest_patch id: :a
    artifact_latest_patch id: :a
    artifact_latest_patch id: :b

    artifact_reset id: :a
    artifact_reset id: :b

    git_artifact_init '/dest', name: 'a', id: :a
    git_artifact_init '/dest_2', name: 'b', id: :b
    artifact_latest_patch id: :a
    artifact_latest_patch id: :b
  end

  it '#remove_latest_patch_if_no_more_diff', test_construct: true do
    git_artifact_init '/dest', changedata: 'text'
    artifact_archive
    artifact_latest_patch
    artifact_latest_patch changedata: 'text', should_be_empty: true

    3.times do |i|
      artifact_layer_patch i + 1, changedata: "text_#{i}"
      artifact_latest_patch
      artifact_latest_patch changedata: "text_#{i}", should_be_empty: true
    end
  end

  { cwd: 'x', paths: 'x', owner: 70_500, group: 70_500 }.each do |param, value|
    it "#autocleanup_on_#{param}_change", test_construct: true do
      artifact_do_test '/dest', layers: 2

      artifact_reset
      git_artifact_init '/dest', **{ param => value }
      artifact_expect_clean
    end
  end

  class << self
    def users_and_groups_to_test
      users = [nil, 'root', 100_500]
      users << 'some_unknown' unless shellout('lsb_release -cs').stdout.strip == 'precise'
      users.product(users)
    end
  end

  users_and_groups_to_test.each do |owner, group|
    it "#change_owner_to_#{owner}_and_group_to_#{group}", test_construct: true do
      artifact_do_test '/dest', owner: owner, group: group
    end
  end

  it '#interlayer_period', test_construct: true do
    artifact_do_test '/dest', interlayer_period: 10
  end

  it '#flush_cache', test_construct: true do
    artifact_do_test '/dest'
    artifact_reset
    git_artifact_init '/dest', flush_cache: true
    artifact_expect_clean
  end

  it '#branch', test_construct: true do
    repo_create_branch 'not_master'

    git_artifact_init '/dest', branch: 'not_master', changedata: 'text'
    artifact_archive
    repo_change_and_commit branch: 'master'
    artifact_latest_patch changedata: 'text', should_be_empty: true

    3.times do |i|
      artifact_layer_patch i + 1, changedata: "text_#{i}"
      repo_change_and_commit branch: 'master'
      artifact_latest_patch changedata: "text_#{i}", should_be_empty: true
    end
  end

  it '#commit_by_step', test_construct: true do
    git_artifact_init '/dest'

    artifact_archive
    expect(artifact.commit_by_step(:prepare)).to eq(artifact.commit_by_step(:build))
    expect(artifact.commit_by_step(:build)).to eq(artifact.commit_by_step(:setup))

    artifact_latest_patch changedata: 'text'
    expect(artifact.commit_by_step(:build)).to eq(artifact.commit_by_step(:prepare))
    expect(artifact.commit_by_step(:setup)).not_to eq(artifact.commit_by_step(:build))

    artifact_layer_patch 1
    expect(artifact.commit_by_step(:build)).not_to eq(artifact.commit_by_step(:prepare))
    expect(artifact.commit_by_step(:setup)).to eq(artifact.commit_by_step(:build))

    artifact_latest_patch changedata: 'text'
    expect(artifact.commit_by_step(:build)).not_to eq(artifact.commit_by_step(:prepare))
    expect(artifact.commit_by_step(:setup)).not_to eq(artifact.commit_by_step(:build))
    expect(artifact.commit_by_step(:prepare)).not_to eq(artifact.commit_by_step(:setup))
  end

  it '#exist_in_step', test_construct: true do
    git_artifact_init '/dest', changefile: 'data1.txt'
    artifact_archive
    expect(artifact.exist_in_step?('data1.txt', :prepare)).to be_truthy

    artifact_latest_patch changefile: 'data2.txt'
    expect(artifact.exist_in_step?('data1.txt', :setup)).to be_truthy
    expect(artifact.exist_in_step?('data2.txt', :setup)).to be_truthy
    expect(artifact.exist_in_step?('data2.txt', :build)).to be_falsy
    expect(artifact.exist_in_step?('data2.txt', :prepare)).to be_falsy

    artifact_layer_patch 1, changefile: 'data3.txt'
    expect(artifact.exist_in_step?('data1.txt', :build)).to be_truthy
    expect(artifact.exist_in_step?('data2.txt', :build)).to be_truthy
    expect(artifact.exist_in_step?('data3.txt', :build)).to be_truthy
    expect(artifact.exist_in_step?('data3.txt', :prepare)).to be_falsy

    artifact_latest_patch changefile: 'data4.txt'
    expect(artifact.exist_in_step?('data1.txt', :setup)).to be_truthy
    expect(artifact.exist_in_step?('data2.txt', :setup)).to be_truthy
    expect(artifact.exist_in_step?('data3.txt', :setup)).to be_truthy
    expect(artifact.exist_in_step?('data4.txt', :setup)).to be_truthy
    expect(artifact.exist_in_step?('data4.txt', :build)).to be_falsy
    expect(artifact.exist_in_step?('data4.txt', :prepare)).to be_falsy
  end

  it '#cwd', test_construct: true do
    git_artifact_init '/dest', cwd: 'a', changefile: 'a/data.txt'
    artifact_archive
    expect(artifact_tar_files).to match_array('data.txt')

    artifact_latest_patch should_be_empty: true
    artifact_layer_patch 1, should_be_empty: true

    artifact_latest_patch changefile: 'a/data.txt'
    artifact_layer_patch 1, changefile: 'a/data.txt'

    artifact_latest_patch should_be_empty: true
    artifact_latest_patch changefile: 'a/data.txt'
  end

  it '#paths', test_construct: true do
    repo_change_and_commit changefile: 'x/data.txt'
    repo_change_and_commit changefile: 'x/y/data.txt'
    repo_change_and_commit changefile: 'z/data.txt'
    git_artifact_init '/dest', paths: ['x/y', 'z']

    artifact_archive
    expect(artifact_tar_files).to match_array(['x/y/data.txt', 'z/data.txt'])

    artifact_latest_patch should_be_empty: true
    repo_change_and_commit changefile: 'x/data.txt'
    artifact_latest_patch should_be_empty: true
    artifact_layer_patch 1, should_be_empty: true
    repo_change_and_commit changefile: 'x/data.txt'

    artifact_latest_patch changefile: 'x/y/data.txt'
    artifact_layer_patch 1, changefile: 'z/data.txt'

    artifact_latest_patch should_be_empty: true
    repo_change_and_commit changefile: 'x/data.txt'
    artifact_latest_patch should_be_empty: true
    artifact_latest_patch changefile: 'x/y/data.txt'
  end

  it '#cwd_and_paths', test_construct: true do
    repo_change_and_commit changefile: 'a/data.txt'
    repo_change_and_commit changefile: 'a/x/data.txt'
    repo_change_and_commit changefile: 'a/x/y/data.txt'
    repo_change_and_commit changefile: 'a/z/data.txt'
    git_artifact_init '/dest', cwd: 'a', paths: ['x/y', 'z']

    artifact_archive
    expect(artifact_tar_files).to match_array(['x/y/data.txt', 'z/data.txt'])

    artifact_latest_patch should_be_empty: true
    repo_change_and_commit changefile: 'a/data.txt'
    artifact_latest_patch should_be_empty: true
    repo_change_and_commit changefile: 'a/x/data.txt'
    artifact_latest_patch should_be_empty: true
    artifact_layer_patch 1, should_be_empty: true
    repo_change_and_commit changefile: 'a/data.txt'
    artifact_layer_patch 1, should_be_empty: true
    repo_change_and_commit changefile: 'a/x/data.txt'

    artifact_latest_patch changefile: 'a/x/y/data.txt'
    artifact_layer_patch 1, changefile: 'a/z/data.txt'

    artifact_latest_patch should_be_empty: true
    repo_change_and_commit changefile: 'a/data.txt'
    artifact_latest_patch should_be_empty: true
    repo_change_and_commit changefile: 'a/x/data.txt'
    artifact_latest_patch should_be_empty: true
    artifact_latest_patch changefile: 'a/x/y/data.txt'
  end

  it '#file_removal_in_patch', test_construct: true do
    git_artifact_init '/dest', changedata: 'test'
    repo_change_and_commit changefile: 'data2.txt', changedata: 'test'

    artifact_archive
    FileUtils.rm File.join(repo.name, 'data2.txt')

    artifact_latest_patch changedata: 'test'
    expect(shellout("zcat #{artifact_filename('_latest.patch.gz')}").stdout).to match(%r{^\+\+\+ /dev/null$})
  end
end
