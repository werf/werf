require_relative '../spec_helper'

describe Dapp::Dimg::Dimg do
  include SpecHelper::Common
  include SpecHelper::Dimg
  include SpecHelper::Git
  
  def projects_origins
    {
      test: 'https://github.com/flant/dapp-submodule-spec-project-test.git',
      https: 'https://github.com/flant/dapp-submodule-spec-project-https.git',
      relative_path: 'https://github.com/flant/dapp-submodule-spec-project-relative-path.git',
      nested: 'https://github.com/flant/dapp-submodule-spec-project-nested.git',
    }
  end

  before :all do
    @project_test_tmpdir = Dir.mktmpdir
    @project_test_path   = File.join(@project_test_tmpdir, 'dapp-submodule-spec-project-test')
    git_clone(projects_origins[:test], @project_test_path)

    [:master, :feature].each do |branch|
      git_checkout(branch, git_dir: @project_test_path)
      git_submodule_update(git_dir: @project_test_path)
      (@project_test_md5sum_by_branch ||= {})[branch] = calculate_md5sum(@project_test_path)
    end

    git_checkout(:master, git_dir: @project_test_path)
  end

  after :all do
    FileUtils.rmtree(@project_test_tmpdir)
  end

  def git_clone(url, path = '.')
    Rugged::Repository.clone_at(url, path)
  end

  def git_checkout(branch, git_dir: '.')
    shellout!("git -C #{git_dir} checkout #{branch}")
  end

  def git_submodule_update(git_dir: '.')
    shellout!("git -C #{git_dir} submodule init")
    shellout!("git -C #{git_dir} submodule update --remote --recursive")
  end

  def calculate_md5sum(path)
    shellout!(md5sum_command(path, ignore_git: true))
      .stdout
      .strip
  end

  def md5sum_command(directory, ignore_git: false)
    shellout_pack_wrapper("find #{directory} -xtype f#{' -not -path "**/.git" -not -path "**/.git/*"' if ignore_git} | xargs md5sum | awk '{ print $1 }' | sort | md5sum | awk '{ print $1 }'")
  end

  def shellout_pack_wrapper(*cmds)
    "bash -ec '#{shellout_pack(cmds.join(' && '))}'"
  end

  def project_test_md5sum(branch = :master)
    @project_test_md5sum_by_branch[branch]
  end

  def clone_project_and_expect_submodule(url, expected_md5sum)
    git_clone(url)
    expect_submodule(expected_md5sum)
  end

  def expect_submodule(expected_md5sum)
    expect { dimg.build! }.to_not raise_error
    expect(git_submodules_paths.count).to eq 1
    expect(container_submodule_md5sum(git_submodules_paths.first)).to eq expected_md5sum
  end

  def git_submodules_paths
    git_repo
      .submodules
      .map(&:path)
  end

  def container_submodule_md5sum(submodule)
    cmd = "#{host_docker} run --rm #{dimg.send(:last_stage).image.name} #{md5sum_command(File.join(local_artifact_to, submodule))}"
    expect { return shellout!(cmd).stdout.strip }.to_not raise_error
  end

  def local_artifact_to
    config._git_artifact._local.first._artifact_options[:to]
  end

  def dapp
    Dapp::Dapp.new(options: dapp_options)
  end

  def config
    dapp.build_configs.first
  end

  def init_dapp_project
    git_init
    dappfile_content = <<EOF
dimg do
docker.from 'ubuntu:16.04'
git.add.to('/app')
end
EOF
    git_change_and_commit('Dappfile', dappfile_content)
  end

  def git_add_and_commit_submodule(url, name)
    git_repo.submodules.add(url, name)
    git_commit
  end

  it 'local', test_construct: true do
    init_dapp_project
    git_add_and_commit_submodule(@project_test_path, 'dapp-submodule-spec-project-test')
    expect_submodule(project_test_md5sum)
  end

  it 'https', test_construct: true do
    clone_project_and_expect_submodule(projects_origins[:https], project_test_md5sum)
  end

  it 'relative_path', test_construct: true do
    clone_project_and_expect_submodule(projects_origins[:relative_path], project_test_md5sum)
  end

  it 'nested', test_construct: true do
    project_https_path = File.join(@project_test_tmpdir, 'dapp-submodule-spec-project-https')
    git_clone(projects_origins[:https], project_https_path)
    git_submodule_update(git_dir: project_https_path)
    project_https_md5sum = calculate_md5sum(project_https_path)

    clone_project_and_expect_submodule(projects_origins[:nested], project_https_md5sum)
  end
end
