require_relative '../spec_helper'

describe Dapp::GitArtifact do
  include SpecHelper::Common
  include SpecHelper::Application
  include SpecHelper::Git
  include SpecHelper::GitArtifact

  before :each do
    stub_application
    stub_docker_image
    stub_git_repo_own

    git_init!
  end

  def config
    default_config.merge(
      _builder: :shell,
      _home_path: '',
      _git_artifact: default_config[:_git_artifact].merge(_local: { _artifact_options: git_artifact_local_options })
    )
  end

  def cli_options
    @cli_options ||= default_cli_options.merge(build: '')
  end

  def git_artifact_local_options
    {
      cwd: (@cwd ||= ''),
      paths: (@paths ||= []),
      branch: (@branch ||= 'master'),
      where_to_add: (@where_to_add ||= '/tmp/dapp-git-artifact-where-to-add'),
      group: (@group ||= 'root'),
      owner: (@owner ||= 'root')
    }
  end

  [:patch, :archive].each do |type|
    define_method "#{type}_apply" do |add_files: ['data.txt'], added_files: add_files, not_added_files: [], **kvargs, &blk|
      [:cwd, :paths, :branch, :where_to_add, :group, :owner].each { |opt| instance_variable_set(:"@#{opt}", kvargs[opt]) unless kvargs[opt].nil? }

      application_build! if type == :patch && !kvargs[:ignore_build]
      add_files.each { |file_path| git_change_and_commit!(file_path, branch: @branch) }
      application_renew

      command_apply(send("#{type}_command"))

      expect(File.exist?(@where_to_add)).to be_truthy
      added_files.each { |file_path| expect(File.exist?(File.join(@where_to_add, file_path))).to be_truthy }
      not_added_files.each { |file_path| expect(File.exist?(File.join(@where_to_add, file_path))).to be_falsey }

      blk.call unless blk.nil? # expectation
      clear_where_to_add
    end
  end

  def archive_command
    git_artifact.archive_apply_command(stages[:source_1_archive])
    stages[:source_1_archive].image.send(:prepared_bash_command)
  end

  def patch_command
    git_artifact.apply_patch_command(stages[:source_5])
    stages[:source_5].image.send(:prepared_bash_command)
  end

  def command_apply(command)
    expect(command).to_not be_empty
    expect { application.shellout!(command) }.to_not raise_error
  end

  def clear_where_to_add
    FileUtils.rm_rf @where_to_add
  end

  def with_credentials(owner_name, group_name)
    shellout! "sudo groupadd #{group_name}; sudo useradd #{owner_name} -g #{group_name}"
    yield
  ensure
    shellout "sudo userdel #{owner_name}"
  end

  def expect_file_credentials(file_path, group_name, user_name)
    file_stat = File.stat(file_path)
    file_group_name = Etc.getgrgid(file_stat.gid).name
    file_user_name  = Etc.getpwuid(file_stat.uid).name
    expect(file_group_name).to eq group_name.to_s
    expect(file_user_name).to eq user_name.to_s
  end

  [:patch, :archive].each do |type|
    it "##{type}", test_construct: true do
      send("#{type}_apply")
    end

    it "##{type} branch", test_construct: true do
      send("#{type}_apply", branch: 'master')
      git_create_branch!('not_master')
      send("#{type}_apply", add_files: ['not_master.txt'], branch: 'not_master')
      send("#{type}_apply", not_added_files: ['not_master.txt'], branch: 'master')
    end

    it "##{type} cwd", test_construct: true do
      send("#{type}_apply", add_files: %w(master.txt a/master2.txt),
                            added_files: ['master2.txt'], not_added_files: %w(a master.txt),
                            cwd: 'a')
    end

    it "##{type} paths", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
                            added_files: %w(x/y/data.txt z/data.txt), not_added_files: ['x/data.txt'],
                            paths: %w(x/y z))
    end

    it "##{type} cwd_and_paths", test_construct: true do
      send("#{type}_apply", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
                            added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(a data.txt),
                            cwd: 'a', paths: %w(x/y z))
    end
  end

  file_name = 'test_file.txt'
  owner = :git_artifact
  group = :git_artifact

  it '#archive owner_and_group', test_construct: true do
    with_credentials(owner, group) do
      archive_apply(add_files: [file_name], owner: owner, group: group) do
        expect_file_credentials(File.join(@where_to_add, file_name), owner, group)
      end
    end
  end

  it '#patch owner_and_group', test_construct: true do
    with_credentials(owner, group) do
      archive_apply(owner: owner, group: group) do
        application_build!
        patch_apply(add_files: [file_name], owner: owner, group: group, ignore_build: true) do
          expect_file_credentials(File.join(@where_to_add, file_name), owner, group)
        end
      end
    end
  end
end
