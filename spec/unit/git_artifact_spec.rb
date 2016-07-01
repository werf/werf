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


  def archive_apply(*args)
    apply(*args) { command_apply(git_artifact.archive_apply_command(stages[:source_1_archive])) }
  end

  def patch_apply(*args)
    apply(*args) { command_apply(git_artifact.apply_patch_command(stages[:source_5])) }
  end

  def apply(add_files: ['data.txt'], added_files: add_files, not_added_files: [], **kvargs)
    [:cwd, :paths, :branch, :where_to_add, :group, :owner].each { |opt| instance_variable_set(:"@#{opt}", kvargs[opt]) unless kvargs[opt].nil? }

    add_files.each { |file_path| git_change_and_commit!(file_path, branch: @branch) }
    application_renew

    yield

    expect(File.exist? @where_to_add).to be_truthy
    added_files.each { |file_path| expect(File.exist? File.join(@where_to_add, file_path)).to be_truthy }
    not_added_files.each { |file_path| expect(File.exist? File.join(@where_to_add, file_path)).to be_falsey }
  end

  def command_apply(command)
    expect(command).to_not be_empty
    expect { application.shellout! command.join(';') }.to_not raise_error
  end

  def clear_where_to_add
    FileUtils.rm_rf @where_to_add
  end

  def before_patch(branch: 'master')
    archive_apply(branch: branch)
    application_build!
  end

  def expect_file_credentials(file_path, group_name, user_name)
    file_stat = File.stat(file_path)
    file_group_name = Etc.getgrgid(file_stat.gid).name
    file_user_name  = Etc.getpwuid(file_stat.uid).name
    expect(file_group_name).to eq group_name
    expect(file_user_name).to eq user_name
  end


  it '#patch', test_construct: true do
    before_patch
    patch_apply
  end

  it '#archive', test_construct: true do
    archive_apply
  end

  it '#archive branch', test_construct: true do
    archive_apply(branch: 'master')
    git_create_branch!('not_master')
    archive_apply(add_files: ['not_master.txt'], branch: 'not_master')
    clear_where_to_add
    archive_apply(branch: 'master', not_added_files: ['not_master.txt'])
  end

  it '#archive cwd', test_construct: true do
    archive_apply(add_files: %w(master.txt a/master2.txt),
                  added_files: ['master2.txt'], not_added_files: ['a', 'master.txt'],
                  cwd: 'a')
  end

  it '#archive paths', test_construct: true do
    archive_apply(add_files: %w(x/data.txt x/y/data.txt z/data.txt),
                  added_files: ['x/y/data.txt', 'z/data.txt'], not_added_files: ['x/data.txt'],
                  paths: %w(x/y z))
  end

  it '#archive cwd_and_paths', test_construct: true do
    archive_apply(add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
                  added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(a data.txt),
                  cwd: 'a', paths: %w(x/y z))
  end

  it '#archive owner_and_group', test_construct: true do
    shellout 'groupadd git_artifact; useradd git_artifact -g git_artifact'

    begin
      archive_apply(add_files: ['test_file.txt'], owner: :git_artifact, group: :git_artifact)
      expect_file_credentials(File.join(@where_to_add, 'test_file.txt'), @group.to_s, @owner.to_s)
    ensure
      shellout 'userdel git_artifact'
    end
  end
end
