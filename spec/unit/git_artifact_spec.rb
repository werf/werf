require_relative '../spec_helper'

describe Dapp::Dimg::GitArtifact do
  include SpecHelper::Common
  include SpecHelper::Git
  include SpecHelper::Dimg

  before :each do
    init_git_artifact_local_options
    git_init
    stub_stages
    stub_dimg
  end

  after :each do
    docker_cleanup
  end

  def init_git_artifact_local_options
    @cwd           = ''
    @to            = '/dist'
    @include_paths = []
    @exclude_paths = []
    @branch        = 'master'
    @group         = 'root'
    @owner         = 'root'
  end

  def stubbed_stage
    instance_double(Dapp::Dimg::Build::Stage::Base).tap do |instance|
      allow(instance).to receive(:prev_stage=)
    end
  end

  def stub_stages
    @stage_commit = {}
    [Dapp::Dimg::Build::Stage::GAArchive, Dapp::Dimg::Build::Stage::GALatestPatch].each do |stage|
      allow_any_instance_of(stage).to receive(:layer_commit) do
        @stage_commit[stage.name] ||= {}
        @stage_commit[stage.name][@branch] ||= git_latest_commit(branch: @branch)
      end
    end
    allow_any_instance_of(Dapp::Dimg::Build::Stage::GALatestPatch).to receive(:prev_stage) { g_a_archive_stage }
  end

  def stub_image
    instance_double(Dapp::Dimg::Image::Stage).tap do |obj|
      allow(obj).to receive(:labels) do
        obj.instance_variable_get(:@_mock_labels) ||
          obj.instance_variable_set(:@_mock_labels, Hash.new('directory'))
      end

      allow(obj).to receive(:add_service_change_label) do |**options|
        options.each do |k, v|
          obj.labels[k.to_s] = v
        end
      end
    end
  end

  def image_build(*cmds)
    container_run(*cmds, rm: false).tap do
      shellout("#{host_docker} commit #{containter_name} #{containter_name}:latest")
      shellout("#{host_docker} rm #{containter_name}")
      @spec_image_name = "#{containter_name}:latest"
    end
  end

  def container_run(*cmds, rm: true)
    shellout(["#{host_docker} run",
              ('--rm' if rm).to_s,
              "--entrypoint #{g_a_stubbed_dimg.dapp.bash_bin}",
              "--name #{containter_name}",
              "--volume #{g_a_stubbed_dimg.tmp_path('archives')}:#{g_a_stubbed_dimg.container_tmp_path('archives')}:ro",
              "--volume #{g_a_stubbed_dimg.tmp_path('patches')}:#{g_a_stubbed_dimg.container_tmp_path('patches')}:ro",
              "--volumes-from #{g_a_stubbed_dimg.dapp.gitartifact_container}",
              "--volumes-from #{g_a_stubbed_dimg.dapp.toolchain_container}",
              "--volumes-from #{g_a_stubbed_dimg.dapp.base_container}",
              "--label dapp=#{g_a_stubbed_dimg.dapp.name}",
              '--label dapp-test=true',
              image_name.to_s,
              prepare_container_command(*cmds).to_s].join(' '))
  end

  def prepare_container_command(*cmds)
    "-ec '#{g_a_stubbed_dimg.dapp.shellout_pack cmds.join(' && ')}'"
  end

  def docker_cleanup
    g_a_stubbed_dimg.dapp.send(:dapp_containers_flush_by_label, 'dapp-test')
    g_a_stubbed_dimg.dapp.send(:dapp_images_flush_by_label, 'dapp-test')
  end

  def containter_name
    @spec_containter_name ||= SecureRandom.uuid
  end

  def image_name
    @spec_image_name ||= reset_image_name
  end

  def reset_image_name
    @spec_image_name = 'ubuntu:16.04'
  end

  # def g_a_stubbed_dapp
  #   instance_double(Dapp::Dapp).tap do |obj|
  #     allow(obj).to receive(:dev_mode?).and_return(false)
  #   end
  # end

  def g_a_stubbed_dimg
    @g_a_stubbed_dimg ||= Dapp::Dimg::Dimg.new(dapp: dapp, config: openstruct_config).tap do |obj|
      orig_method = obj.method(:stage_by_name)

      allow(obj).to receive(:stage_by_name) do |name|
        if name == :g_a_archive
          g_a_archive_stage
        else
          orig_method.call(name)
        end
      end
    end
  end

  def g_a_archive_stage
    @g_a_archive_stage ||= begin
      Dapp::Dimg::Build::Stage::GAArchive.new(g_a_stubbed_dimg, stubbed_stage).tap do |obj|
        allow(obj).to receive(:image) do
          obj.instance_variable_get(:@_stub_image) ||
            obj.instance_variable_set(:@_stub_image, stub_image)
        end
      end
    end
  end

  def g_a_latest_patch_stage
    @g_a_latest_patch_stage ||= Dapp::Dimg::Build::Stage::GALatestPatch.new(g_a_stubbed_dimg, stubbed_stage)
  end

  def git_artifact
    Dapp::Dimg::GitArtifact.new(stubbed_repo, **git_artifact_local_options)
  end

  def stubbed_repo
    @stubbed_repo ||= Dapp::Dimg::GitRepo::Own.new(g_a_stubbed_dimg)
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
    apply_command(*git_artifact.apply_archive_command(g_a_archive_stage))
  end

  def apply_patch
    apply_command(*git_artifact.apply_patch_command(g_a_latest_patch_stage))
  end

  def apply_command(*cmds)
    image_build(*cmds).tap do |res|
      expect { res.error! }.to_not raise_error
    end
  end

  def expect_existing_container_file(path)
    expect { check_container_file(path).error! }.to_not raise_error
  end

  def expect_not_existing_container_file(path)
    expect { check_container_file(path).error! }.to raise_error ::Mixlib::ShellOut::ShellCommandFailed
  end

  def check_container_file(path)
    container_run("#{g_a_stubbed_dimg.dapp.test_bin} -f \"#{path}\"")
  end

  def container_file_path(path)
    File.join(@to, path)
  end

  def container_file_stat(path)
    res = container_run("#{g_a_stubbed_dimg.dapp.stat_bin} -c '%a %u %g' \"#{path}\"")
    expect { res.error! }.to_not raise_error
    mode, uid, gid = res.stdout.strip.split
    { mode: mode, uid: uid, gid: gid }
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

      added_files.each { |file_path| expect_existing_container_file(container_file_path(file_path)) }
      not_added_files.each { |file_path| expect_not_existing_container_file(container_file_path(file_path)) }
    end

    def reset_image
      reset_image_name
    end

    [:patch, :archive].each do |type|
      it type.to_s, test_construct: true do
        send("check_#{type}")
      end

      it "#{type} branch", test_construct: true do
        send("check_#{type}", branch: 'master')
        reset_image
        send("check_#{type}", add_files: ['not_master.txt'], branch: 'not_master')
        reset_image
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

      it "#{type} paths (glob *.rb)", test_construct: true do
        send("check_#{type}", add_files: %w(data.rb data.txt x/data.rb),
             added_files: %w(data.rb), not_added_files: %w(data.txt x/data.rb),
             include_paths: %w(*.rb))
      end

      it "#{type} paths (glob **/*.rb)", test_construct: true do
        send("check_#{type}", add_files: %w(data.rb x/data.rb z/data.txt),
             added_files: %w(x/data.rb data.rb), not_added_files: %w(z/data.txt),
             include_paths: %w(**/*.rb))
      end

      it "#{type} paths (glob **/dir/**/*)", test_construct: true do
        send("check_#{type}", add_files: %w(dir/data.rb dir/data/data.rb a/dir.txt b/dir/data.rb b/dir/data.txt),
             added_files: %w(dir/data.rb dir/data/data.rb b/dir/data.rb), not_added_files: %w(a/dir.txt b/dir/data.txt),
             include_paths: %w(**/dir/**/*.rb))
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

      it "#{type} exclude_paths (glob *.rb)", test_construct: true do
        send("check_#{type}", add_files: %w(data.rb data.txt x/data.rb),
             added_files: %w(data.txt x/data.rb), not_added_files: %w(data.rb),
             exclude_paths: %w(*.rb))
      end

      it "#{type} exclude_paths (glob **/*.rb)", test_construct: true do
        send("check_#{type}", add_files: %w(data.rb x/data.rb z/data.txt),
             added_files: %w(z/data.txt), not_added_files: %w(data.rb x/data.rb),
             exclude_paths: %w(**/*.rb))
      end

      it "#{type} exclude_paths (glob **/dir/**/*)", test_construct: true do
        send("check_#{type}", add_files: %w(dir/data.rb dir/data/data.rb a/dir.txt b/dir/data.rb b/dir/data.txt),
             added_files: %w(a/dir.txt b/dir/data.txt), not_added_files: %w(dir/data.rb dir/data/data.rb b/dir/data.rb),
             exclude_paths: %w(**/dir/**/*.rb))
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

      it "#{type} file with spaces in name", test_construct: true do
        send("check_#{type}", add_files: ['name with spaces'], added_files: ['name with spaces'])
      end
    end

    context 'owner and group' do
      def expect_container_file_credentials(path, uid, gid)
        file_stat = container_file_stat(path)
        expect(file_stat[:uid]).to eq uid
        expect(file_stat[:gid]).to eq gid
      end

      file_name = 'test_file.txt'
      uid = '1111'
      gid = '1111'

      it 'archive owner_and_group', test_construct: true do
        check_archive(add_files: [file_name], owner: uid, group: gid)
        expect_container_file_credentials(container_file_path(file_name), uid, gid)
      end

      it 'patch owner_and_group', test_construct: true do
        check_archive(owner: uid, group: gid)
        check_patch(add_files: [file_name], owner: uid, group: gid, ignore_init_build: true)
        expect_container_file_credentials(container_file_path(file_name), uid, gid)
      end
    end
  end

  context 'cycle with cwd' do
    def expect_container_file_mode(path, mode)
      expect(mode).to eq container_file_stat(path)[:mode]
    end

    def change_file_mode(path)
      file_mode = File.stat(path).mode
      available_permissions = { 0o100644 => '644', 0o100755 => '755' }
      permission = available_permissions.keys[available_permissions.keys.index(file_mode) - 1]
      File.chmod(permission, path)
      available_permissions[permission]
    end

    [false, true].each do |binary|
      context binary ? 'binary file' : 'file' do
        file_path = 'a/data'
        file_path_without_cwd = 'data'

        before :each do
          git_change_and_commit(file_path, binary: binary)
          @cwd = 'a'
          apply_archive
        end

        it 'added', test_construct: true do
          git_change_and_commit('a/data2', binary: binary)
          apply_patch
        end

        it 'modified', test_construct: true do
          git_change_and_commit(file_path, binary: binary)
          apply_patch
        end

        it 'change_mode', test_construct: true do
          expected_permission = change_file_mode(file_path)
          git_add_and_commit(file_path)
          apply_patch
          expect_container_file_mode(container_file_path(file_path_without_cwd), expected_permission)
        end

        it 'modified and change mode', test_construct: true do
          expected_permission = change_file_mode(file_path)
          git_change_and_commit(file_path, binary: binary)
          apply_patch
          expect_container_file_mode(container_file_path(file_path_without_cwd), expected_permission)
        end

        it 'delete', test_construct: true do
          FileUtils.rm_rf file_path
          git_rm_and_commit file_path
          apply_patch
          expect_not_existing_container_file(container_file_path(file_path_without_cwd))
        end
      end
    end

    context 'symlink' do
      def check_container_symlink(path)
        container_run("#{g_a_stubbed_dimg.dapp.test_bin} -h #{path}")
      end

      def expext_container_file_link(path, link)
        expect { check_container_symlink(path).error! }.to_not raise_error
        expect(container_run("#{g_a_stubbed_dimg.dapp.readlink_bin} -f #{path}").stdout.strip).to eq link
      end

      def git_add_symlink_and_commit(file, link)
        FileUtils.safe_unlink(link) if File.symlink?(link)
        git_change_and_commit(file)
        FileUtils.symlink(file, link)
        git_add_and_commit(link)
      end

      def git_rm_symlink_and_commit(link)
        FileUtils.safe_unlink(link)
        git_rm_and_commit link
      end

      symlink_file = 'symlink'
      symlink2_file = 'symlink2'

      before :each do
        git_add_symlink_and_commit('test', symlink_file)
        apply_archive
        expext_container_file_link(container_file_path(symlink_file), container_file_path('test'))
      end

      it 'added', test_construct: true do
        git_add_symlink_and_commit('test2', symlink2_file)
        apply_patch
        expext_container_file_link(container_file_path(symlink2_file), container_file_path('test2'))
      end

      it 'modified', test_construct: true do
        git_add_symlink_and_commit('test2', symlink_file)
        apply_patch
        expext_container_file_link(container_file_path(symlink_file), container_file_path('test2'))
      end

      it 'delete', test_construct: true do
        git_rm_symlink_and_commit(symlink_file)
        apply_patch
        expect { check_container_symlink(container_file_path(symlink_file)).error! }.to raise_error ::Mixlib::ShellOut::ShellCommandFailed
      end
    end
  end
end
