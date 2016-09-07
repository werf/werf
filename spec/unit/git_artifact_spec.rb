require_relative '../spec_helper'

describe Dapp::GitArtifact do
  include SpecHelper::Common
  include SpecHelper::Git
  include SpecHelper::Application

  before :each do
    init
    git_init!
    stub_stages
    stub_application
  end

  def init
    FileUtils.mkdir 'project'
    @where_to_add = File.expand_path('where-to-add')
    Dir.chdir File.expand_path('project')
  end

  def config
    @config ||= default_config.merge(_home_path: '')
  end

  def stubbed_stage
    instance_double(Dapp::Build::Stage::Base).tap do |instance|
      allow(instance).to receive(:prev_stage=)
    end
  end

  def stub_stages
    @stage_commit = {}
    [Dapp::Build::Stage::GAArchive, Dapp::Build::Stage::GALatestPatch].each do |stage|
      allow_any_instance_of(stage).to receive(:layer_commit) do
        @stage_commit[stage.name] ||= {}
        @stage_commit[stage.name][@branch] ||= git_latest_commit(branch: @branch)
      end
    end
    allow_any_instance_of(Dapp::Build::Stage::GALatestPatch).to receive(:prev_g_a_stage) { g_a_archive_stage }
  end

  def project
    super do
      allow_any_instance_of(Dapp::Project).to receive(:git_path) { 'git' }
      allow_any_instance_of(Dapp::Project).to receive(:sudo_path) { 'sudo' }
    end
  end

  def g_a_archive_stage
    @g_a_archive_stage ||= Dapp::Build::Stage::GAArchive.new(empty_application, stubbed_stage)
  end

  def g_a_latest_patch_stage
    @g_a_latest_patch_stage ||= Dapp::Build::Stage::GALatestPatch.new(empty_application, stubbed_stage)
  end

  def git_artifact
    Dapp::GitArtifact.new(stubbed_repo, **git_artifact_local_options)
  end

  def stubbed_repo
    instance_double(Dapp::GitRepo::Own).tap do |instance|
      allow(instance).to receive(:container_path) { '.git' }
      allow(instance).to receive(:application) { application }
    end
  end

  def git_artifact_local_options
    {
      cwd: (@cwd ||= ''),
      paths: (@paths ||= []),
      exclude_paths: (@exclude_paths ||= []),
      branch: (@branch ||= 'master'),
      where_to_add: @where_to_add,
      group: (@group ||= 'root'),
      owner: (@owner ||= 'root')
    }
  end

  [:patch, :archive].each do |type|
    define_method "#{type}_apply" do |add_files: ['data.txt'], added_files: add_files, not_added_files: [], **kwargs, &blk|
      @branch = kwargs[:branch] unless kwargs[:branch].nil?
      command_apply(archive_command) if type == :patch && !kwargs[:ignore_archive_apply]

      [:cwd, :paths, :exclude_paths, :where_to_add, :group, :owner].each { |opt| instance_variable_set(:"@#{opt}", kwargs[opt]) unless kwargs[opt].nil? }
      add_files.each { |file_path| git_change_and_commit!(file_path, branch: @branch) }

      command_apply(send("#{type}_command"))

      expect(File.exist?(@where_to_add)).to be_truthy
      added_files.each { |file_path| expect(File.exist?(File.join(@where_to_add, file_path))).to be_truthy }
      not_added_files.each { |file_path| expect(File.exist?(File.join(@where_to_add, file_path))).to be_falsey }

      blk.call unless blk.nil? # expectation
      remove_where_to_add
    end
  end

  def archive_command
    git_artifact.apply_archive_command(g_a_archive_stage)
  end

  def patch_command
    git_artifact.apply_patch_command(g_a_latest_patch_stage)
  end

  def command_apply(command)
    expect(command).to_not be_empty
    shellout(%(bash -ec '#{command.join(' && ')}')).tap do |res|
      expect { res.error! }.to_not raise_error, res.inspect
    end
  end

  def remove_where_to_add
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
    it type.to_s, test_construct: true do
      send("#{type}_apply")
    end

    it "#{type} branch", test_construct: true do
      send("#{type}_apply", branch: 'master')
      git_create_branch!('not_master')
      send("#{type}_apply", add_files: ['not_master.txt'], branch: 'not_master')
      send("#{type}_apply", not_added_files: ['not_master.txt'], branch: 'master')
    end

    it "#{type} cwd", test_construct: true do
      send("#{type}_apply", add_files: %w(master.txt a/master2.txt),
                            added_files: ['master2.txt'], not_added_files: %w(a master.txt),
                            cwd: 'a')
    end

    it "#{type} paths", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
                            added_files: %w(x/y/data.txt z/data.txt), not_added_files: ['x/data.txt'],
                            paths: %w(x/y z))
    end

    it "#{type} paths (files)", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
           added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(x/data.txt),
           paths: %w(x/y/data.txt z/data.txt))
    end

    it "#{type} paths (globs)", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
           added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(x/data.txt),
           paths: %w(x/y/* z/[asdf]ata.txt))
    end

    it "#{type} cwd and paths", test_construct: true do
      send("#{type}_apply", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
                            added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(a data.txt),
                            cwd: 'a', paths: %w(x/y z))
    end

    it "#{type} exclude_paths", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
           added_files: %w(z/data.txt), not_added_files: %w(x/data.txt x/y/data.txt),
           exclude_paths: %w(x))
    end

    it "#{type} exclude_paths (files)", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
           added_files: %w(x/data.txt), not_added_files: %w(x/y/data.txt z/data.txt),
           exclude_paths: %w(x/y/data.txt z/data.txt))
    end

    it "#{type} exclude_paths (globs)", test_construct: true do
      send("#{type}_apply", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
           added_files: %w(x/data.txt), not_added_files: %w(x/y/data.txt z/data.txt),
           exclude_paths: %w(x/y/* z/[asdf]*ta.txt))
    end

    it "#{type} cwd and exclude_paths", test_construct: true do
      send("#{type}_apply", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
           added_files: %w(data.txt z/data.txt), not_added_files: %w(a x/y/data.txt),
           cwd: 'a', exclude_paths: %w(x))
    end

    it "#{type} cwd, paths and exclude_paths", test_construct: true do
      send("#{type}_apply", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
           added_files: %w(x/data.txt z/data.txt), not_added_files: %w(a data.txt x/y/data.txt),
           cwd: 'a', paths: [%w(x z)], exclude_paths: %w(x/y))
    end
  end

  file_name = 'test_file.txt'
  owner = :git_artifact
  group = :git_artifact

  it 'archive owner_and_group', test_construct: true do
    with_credentials(owner, group) do
      archive_apply(add_files: [file_name], owner: owner, group: group) do
        expect_file_credentials(File.join(@where_to_add, file_name), owner, group)
      end
    end
  end

  it 'patch owner_and_group', test_construct: true do
    with_credentials(owner, group) do
      archive_apply(owner: owner, group: group) do
        patch_apply(add_files: [file_name], owner: owner, group: group, ignore_archive_apply: true) do
          expect_file_credentials(File.join(@where_to_add, file_name), owner, group)
        end
      end
    end
  end
end
