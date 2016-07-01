require_relative '../spec_helper'

describe Dapp::GitArtifact do
  include SpecHelpers::Common
  include SpecHelpers::Application
  include SpecHelpers::Git
  include SpecHelpers::GitArtifact

  before :each do
    stub_application
    stub_docker_image
    stub_git_repo_own

    git_init!
  end


  def config
    { from: 'ubuntu:16.04', type: :shell, git_artifact: { local: git_artifact_local_options } }
  end

  def change_artifact_branch(branch = 'master')
    config[:git_artifact][:local][:branch] = branch
  end

  def git_artifact_local_options
    {
      cwd: (@cwd ||= ''),
      paths: (@paths ||= []),
      branch: (@branch ||= 'master'),
      where_to_add: (@where_to_add ||= 'dest'),
      group: (@group ||= 'root'),
      owner: (@owner ||= 'root')
    }
  end


  def archive_apply(add_files: ['archive_data.txt'], added_files: add_files, not_added_files: [], **kwargs)
    [:cwd, :paths, :branch, :where_to_add, :group, :owner].each { |opt| instance_variable_set(:"@#{opt}", kwargs[opt]) unless kwargs[opt].nil? }

    add_files.each { |file_path| git_change_and_commit!(file_path, branch: @branch) }
    application_renew

    command_apply(git_artifact.archive_apply_command(stages[:source_1_archive]))

    expect(File.exist? @where_to_add).to be_truthy
    added_files.each { |file_path| expect(File.exist? File.join(@where_to_add, file_path)).to be_truthy }
    not_added_files.each { |file_path| expect(File.exist? File.join(@where_to_add, file_path)).to be_falsey }
    expect(File.exist? '.tar.gz').to be_falsey
  end

  def patch_apply(patch_file = 'new_file.txt')
    git_change_and_commit!(patch_file)
    application_renew
    command_apply(git_artifact.apply_patch_command(stages[:source_5]))

    expect(File.exist? @where_to_add).to be_truthy
    expect(File.exist? File.join(@where_to_add, patch_file)).to be_truthy
  end

  def command_apply(command)
    expect(command).to_not be_empty
    expect { application.shellout! command.join(';') }.to_not raise_error
  end

  def clear_where_to_add
    FileUtils.rm_rf @where_to_add
  end


  it '#archive', test_construct: true do
    archive_apply
  end

  it '#patch', test_construct: true do
    archive_apply
    application_build!
    patch_apply
  end

  it '#branch', test_construct: true do
    archive_apply(branch: 'master')
    git_create_branch!('not_master')
    archive_apply(add_files: ['not_master.txt'], branch: 'not_master')
    clear_where_to_add
    archive_apply(branch: 'master', not_added_files: ['not_master.txt'])
  end

  it '#cwd', test_construct: true do
    archive_apply(add_files: %w(master.txt a/master2.txt),
                  added_files: ['master2.txt'], not_added_files: ['a', 'master.txt'],
                  cwd: 'a')
  end

  it '#paths', test_construct: true do
    archive_apply(add_files: %w(x/data.txt x/y/data.txt z/data.txt),
                  added_files: ['x/y/data.txt', 'z/data.txt'], not_added_files: ['x/data.txt'],
                  paths: %w(x/y z))
  end

  it '#cwd_and_paths', test_construct: true do
    archive_apply(add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
                  added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(a data.txt),
                  cwd: 'a', paths: %w(x/y z))
  end

  xit '#owner_and_group', test_construct: true do
    archive_apply(add_files: ['test_file.txt'], owner: :nobody, group: :nobody)
    expect_file_owner(File.join(@where_to_add, 'test_file.txt'), @owner)
  end

  def expect_file_owner(file_path, owner_name)
    owner = Etc.getgrnam(owner_name)
    expect(File.stat(file_path).id).to eq owner.id
  end
end
