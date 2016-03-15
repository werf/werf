require_relative 'spec_helper'

describe Dapp::GitArtifact do
  before :all do
    shellout 'git config -l | grep "user.email" || git config --global user.email "dapp@flant.com"'
    shellout 'git config -l | grep "user.name" || git config --global user.name "Dapp Dapp"'
  end

  before :each do
    @builder = instance_double('Dapp::Builder')
    allow(@builder).to receive(:register_atomizer)
    allow(@builder).to receive(:build_path) do |*args|
      File.join(*args)
    end
    allow(@builder).to receive(:home_path).and_return('')
    allow(@builder).to receive(:shellout) do |*args, **kwargs|
      shellout(*args, **kwargs)
    end
    allow(@builder).to receive(:filelock).and_yield

    @docker = instance_double('Dapp::Docker')
    allow(@docker).to receive(:add_artifact)
    allow(@docker).to receive(:run)
    allow(@builder).to receive(:docker).and_return(@docker)

    @repo = Dapp::GitRepo::Chronicler.new(@builder, 'repo')
  end

  def reset_instances
    RSpec::Mocks.space.proxy_for(@builder).send(:instance_variable_get, :@messages_received).clear
    RSpec::Mocks.space.proxy_for(@docker).send(:instance_variable_get, :@messages_received).clear
  end

  after :each do
    Timecop.return
  end

  # where to add
  # name
  # branch: 'master'
  # cwd: nil
  # paths: nil
  # owner: nil, group: nil
  # interlayer_period: 7 * 24 * 3600
  # build_path: nil
  # flush_cache: false

  def commit(changefile, changedata, branch: 'master')
    shellout "cd repo; git checkout #{branch}"
    changefile = File.join('repo', changefile)
    FileUtils.mkdir_p File.split(changefile)[0]
    File.write changefile, changedata
    @repo.commit!
  end

  def artifact_init(where_to_add, id: nil, changefile: nil, changedata: random_string, **kwargs)
    commit(changefile, changedata) if changefile

    (@artifact ||= {})[id] = Dapp::GitArtifact.new(@builder, @repo, where_to_add, **kwargs)
  end

  def artifact_reset(id: nil)
    @artifact.delete(id).send(:atomizer).tap do |atomizer|
      atomizer.commit!
      atomizer.send(:instance_variable_get, :@file).close
    end
  end

  def artifact_filename(ending, id: nil)
    "#{@artifact[id].send(:repo).name}#{@artifact[id].send(:name) ? "_#{@artifact[id].send(:name)}" : nil}.#{@artifact[id].send(:branch)}#{ending}"
  end

  def artifact_archive(id: nil)
    reset_instances
    @artifact[id].add_multilayer!

    expect(@docker).to have_received(:add_artifact).with(
      artifact_filename('.tar.gz', id: id),
      artifact_filename('.tar.gz', id: id),
      @artifact[id].send(:where_to_add),
      step: :prepare
    )
    expect(File.read(artifact_filename('.commit', id: id)).strip).to eq(@repo.latest_commit)
    expect(File.exist?(artifact_filename('.tar.gz', id: id))).to be_truthy
  end

  def random_string
    (('a'..'z').to_a * 10).sample(100).join
  end

  def artifact_latest_patch(id: nil, **kwargs)
    artifact_patch(
      '_latest',
      :setup,
      id: id,
      **kwargs
    )
  end

  def artifact_layer_patch(layer, id: nil, **kwargs)
    Timecop.travel(Time.now + @artifact[id].send(:interlayer_period))

    artifact_patch(
      format('_layer_%04d', layer),
      :build,
      id: id,
      **kwargs
    )
  ensure
    Timecop.return
  end

  # rubocop:disable Metrics/AbcSize
  def artifact_patch(suffix, step, id:, changefile: 'data.txt', changedata: random_string)
    commit(changefile, changedata)

    reset_instances
    @artifact[id].add_multilayer!

    patch_filename = artifact_filename("#{suffix}.patch.gz", id: id)
    commit_filename = artifact_filename("#{suffix}.commit", id: id)

    expect(@docker).to have_received(:add_artifact).with(patch_filename, patch_filename, '/tmp', step: step)
    expect(@docker).to have_received(:run).with(
      "zcat /tmp/#{patch_filename} | git apply --whitespace=nowarn --directory=#{@artifact[id].send(:where_to_add)}",
      "rm /tmp/#{patch_filename}",
      step: step
    )
    expect(File.read(commit_filename).strip).to eq(@repo.latest_commit)
    expect(File.exist?(patch_filename)).to be_truthy
    expect(shellout("zcat #{patch_filename}").stdout).to match(/#{changedata}/)
  end
  # rubocop:enable Metrics/AbcSize

  def artifact_do_test(where_to_add, latest_patch: true, layers: 3)
    artifact_init where_to_add
    artifact_archive
    layers.times do |i|
      artifact_layer_patch i + 1
    end
    artifact_latest_patch if latest_patch
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
    artifact_init '/dest', name: 'a', id: :a
    artifact_init '/dest_2', name: 'b', id: :b
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

    artifact_init '/dest', name: 'a', id: :a
    artifact_init '/dest_2', name: 'b', id: :b
    artifact_latest_patch id: :a
    artifact_latest_patch id: :b
  end

  it '#no_patch_if_no_more_diff', test_construct: true do
  end

  { cwd: 'x', paths: 'x', owner: 70_500, group: 70_500 }.each do |param, value|
    it "#autocleanup_on_#{param}_change", test_construct: true do
      artifact_do_test '/dest', layers: 2

      artifact_reset
      artifact_init '/dest', **{ param => value }
      expect(Dir.glob(artifact_filename('{.,_}*'))).to eq([artifact_filename('.paramshash'), artifact_filename('.atomizer')])
    end
  end
end
