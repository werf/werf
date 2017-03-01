require_relative '../spec_helper'

describe Dapp::GitArtifact do
  include SpecHelper::Common
  include SpecHelper::Git
  include SpecHelper::Dimg

  before :each do
    init
    git_init
    stub_stages
    stub_dimg
  end

  def init
    FileUtils.mkdir 'project'
    @to = File.expand_path('dist')
    Dir.chdir File.expand_path('project')

    init_git_artifact_local_options
  end

  def init_git_artifact_local_options
    @cwd           = ''
    @include_paths = []
    @exclude_paths = []
    @branch        = 'master'
    @group         = 'root'
    @owner         = 'root'
  end

  def stubbed_stage
    instance_double(Dapp::Build::Stage::Base).tap do |instance|
      allow(instance).to receive(:prev_stage=)
    end
  end

  def stub_dimg
    super do |instance|
      allow(instance).to receive(:container_tmp_path) { |*m_args| instance.tmp_path(*m_args) }
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
      allow_any_instance_of(Dapp::Project).to receive(:git_bin) { 'git' }
      allow_any_instance_of(Dapp::Project).to receive(:tar_bin) { 'tar' }
      allow_any_instance_of(Dapp::Project).to receive(:sudo_bin) { 'sudo' }
      allow_any_instance_of(Dapp::Project).to receive(:install_bin) { 'install' }
    end
  end

  def g_a_archive_stage
    @g_a_archive_stage ||= Dapp::Build::Stage::GAArchive.new(empty_dimg, stubbed_stage)
  end

  def g_a_latest_patch_stage
    @g_a_latest_patch_stage ||= Dapp::Build::Stage::GALatestPatch.new(empty_dimg, stubbed_stage)
  end

  def git_artifact
    Dapp::GitArtifact.new(stubbed_repo, **git_artifact_local_options)
  end

  def stubbed_repo
    @stubbed_repo ||= begin
      Dapp::GitRepo::Own.new(dimg)
    end
  end

  def git_artifact_local_options
    {
      cwd:           @cwd,
      include_paths: @include_paths,
      exclude_paths: @exclude_paths,
      branch:        @branch,
      to:            @to,
      group:         @group,
      owner:         @owner
    }
  end

  def git_change_and_commit(*args, branch: nil, git_dir: '.', **kwargs)
    git_checkout(branch, git_dir: git_dir) unless branch.nil?
    super(*args, git_dir: git_dir, **kwargs)
  end

  def apply_archive
    apply_command(git_artifact.apply_archive_command(g_a_archive_stage))
  end

  def apply_patch
    apply_command(git_artifact.apply_patch_command(g_a_latest_patch_stage))
  end

  def apply_command(command)
    command = Array(command)
    shellout(%(bash -ec '#{command.join(' && ')}')).tap do |res|
      expect { res.error! }.to_not raise_error, res.inspect
    end
  end

  def file_exist_at_dist?(file_path)
    File.exist?(File.join(@to, file_path))
  end

  context 'base' do
    def check_archive(**kwargs)
      git_create_branch(kwargs[:branch]) unless kwargs[:branch].nil?
      check_base(:archive, **kwargs)
    end

    def check_patch(ignore_init_build: false, add_files: [], added_files: add_files, not_added_files: [], **kwargs)
      check_archive(**kwargs) unless ignore_init_build
      check_base(:patch, add_files: add_files, added_files: added_files, not_added_files: not_added_files, **kwargs)
    end

    def check_base(type, add_files: [], added_files: add_files, not_added_files: [], **kwargs)
      [:cwd, :include_paths, :exclude_paths, :to, :group, :owner, :branch].each do |opt|
        instance_variable_set(:"@#{opt}", kwargs[opt]) unless kwargs[opt].nil?
      end

      add_files.each { |file_path| git_change_and_commit(file_path, branch: @branch) }

      send("apply_#{type}")

      expect(File.exist?(@to)).to be_truthy
      added_files.each { |file_path| expect(file_exist_at_dist?(file_path)).to be_truthy }
      not_added_files.each { |file_path| expect(file_exist_at_dist?(file_path)).to be_falsey }

      yield if block_given?
    end

    def cleanup_dist
      FileUtils.rm_rf @to
    end

    [:patch, :archive].each do |type|
      it type.to_s, test_construct: true do
        send("check_#{type}")
      end

      it "#{type} branch", test_construct: true do
        send("check_#{type}", branch: 'master')
        cleanup_dist
        send("check_#{type}", add_files: ['not_master.txt'], branch: 'not_master')
        cleanup_dist
        send("check_#{type}", not_added_files: ['not_master.txt'], branch: 'master')
      end

      it "#{type} cwd", test_construct: true do
        send("check_#{type}", add_files: %w(master.txt a/master2.txt),
             added_files: ['master2.txt'], not_added_files: %w(a master.txt),
             cwd: 'a')
      end

      it "#{type} paths", test_construct: true do
        send("check_#{type}", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
             added_files: %w(x/y/data.txt z/data.txt), not_added_files: ['x/data.txt'],
             include_paths: %w(x/y z))
      end

      it "#{type} paths (files)", test_construct: true do
        send("check_#{type}", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
             added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(x/data.txt),
             include_paths: %w(x/y/data.txt z/data.txt))
      end

      it "#{type} paths (globs)", test_construct: true do
        send("check_#{type}", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
             added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(x/data.txt),
             include_paths: %w(x/y/* z/[asdf]ata.txt))
      end

      it "#{type} (file doesn't exist)", test_construct: true do
        send("check_#{type}", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
             added_files: [], not_added_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
             cwd: 'a/x/c')
      end

      it "#{type} cwd and paths", test_construct: true do
        send("check_#{type}", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
             added_files: %w(x/y/data.txt z/data.txt), not_added_files: %w(a data.txt),
             cwd: 'a', include_paths: %w(x/y z))
      end

      it "#{type} exclude_paths", test_construct: true do
        send("check_#{type}", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
             added_files: %w(z/data.txt), not_added_files: %w(x/data.txt x/y/data.txt),
             exclude_paths: %w(x))
      end

      it "#{type} exclude_paths (files)", test_construct: true do
        send("check_#{type}", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
             added_files: %w(x/data.txt), not_added_files: %w(x/y/data.txt z/data.txt),
             exclude_paths: %w(x/y/data.txt z/data.txt))
      end

      it "#{type} exclude_paths (globs)", test_construct: true do
        send("check_#{type}", add_files: %w(x/data.txt x/y/data.txt z/data.txt),
             added_files: %w(x/data.txt), not_added_files: %w(x/y/data.txt z/data.txt),
             exclude_paths: %w(x/y/* z/[asdf]*ta.txt))
      end

      it "#{type} cwd and exclude_paths", test_construct: true do
        send("check_#{type}", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
             added_files: %w(data.txt z/data.txt), not_added_files: %w(a x/y/data.txt),
             cwd: 'a', exclude_paths: %w(x))
      end

      it "#{type} cwd, paths and exclude_paths", test_construct: true do
        send("check_#{type}", add_files: %w(a/data.txt a/x/data.txt a/x/y/data.txt a/z/data.txt),
             added_files: %w(x/data.txt z/data.txt), not_added_files: %w(a data.txt x/y/data.txt),
             cwd: 'a', include_paths: [%w(x z)], exclude_paths: %w(x/y))
      end
    end

    xcontext 'owner and group' do
      def with_credentials(owner_name, group_name, uid, gid)
        shellout "groupadd #{group_name} --gid #{gid}"
        shellout "useradd #{owner_name} --gid #{gid} --uid #{uid}"
        yield
      ensure
        shellout "userdel #{owner_name}"
      end

      def expect_file_credentials(file_path, uid, gid)
        file_stat = File.stat(file_path)
        expect(file_stat.uid).to eq uid
        expect(file_stat.gid).to eq gid
      end

      file_name = 'test_file.txt'
      owner = :dapp_git_artifact
      group = :dapp_git_artifact
      uid = 1100 + (rand * 1000).to_i
      gid = 1100 + (rand * 1000).to_i

      xit 'archive owner_and_group', test_construct: true do
        with_credentials(owner, group, uid, gid) do
          check_archive(add_files: [file_name], owner: owner, group: group) do
            expect_file_credentials(File.join(@to, file_name), uid, gid)
          end
        end
      end

      xit 'patch owner_and_group', test_construct: true do
        with_credentials(owner, group, uid, gid) do
          check_archive(owner: owner, group: group) do
            check_patch(add_files: [file_name], owner: owner, group: group, ignore_init_build: true) do
              expect_file_credentials(File.join(@to, file_name), uid, gid)
            end
          end
        end
      end
    end
  end

  context 'file cycle with cwd' do
    def file_change_mode(file_path)
      file_mode = File.stat(file_path).mode
      available_permissions = [0100644, 0100755]

      available_permissions[available_permissions.index(file_mode) - 1].tap do |permission|
        File.chmod(permission, file_path)
      end
    end

    file_path = 'a/data.txt'

    before :each do
      git_change_and_commit(file_path)
      @cwd = 'a'
      apply_archive
    end

    it 'modified', test_construct: true do
      git_change_and_commit(file_path)
      apply_patch
    end

    it 'change_mode', test_construct: true do
      expected_permission = file_change_mode(file_path)
      git_add_and_commit(file_path)
      apply_patch
      expect(File.stat(file_path).mode).to eq expected_permission
    end

    it 'modified and change mode', test_construct: true do
      expected_permission = file_change_mode(file_path)
      git_change_and_commit(file_path)
      apply_patch
      expect(File.stat(file_path).mode).to eq expected_permission
    end

    it 'delete', test_construct: true do
      FileUtils.rm_rf file_path
      git_rm_and_commit file_path
      apply_patch
      expect(file_exist_at_dist?(file_path)).to be_falsey
    end
  end
end
