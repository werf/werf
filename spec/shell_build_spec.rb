require_relative 'spec_helper'

describe Dapp::Build::Shell do
  it 'cache' do
    init_repo
    build_run
    modifiable_stages.reverse_each { |s| send(s) }
  end

  # build

  def build_run
    build.run
  end

  def build
    options = { builder: builder, conf: config.dup, opts: opts }
    Dapp::Build::Shell.new(**options).tap { |build| build.docker = docker }
  end

  def builder
    instance_double('Dapp::Builder').tap do |obj|
      allow(obj).to receive(:register_docker_atomizer).and_return([])
      allow(obj).to receive(:register_file_atomizer).and_return begin
                                                                  instance_double('Dapp::Atomizer::File').tap do |atomizer|
                                                                    allow(atomizer).to receive(:add_path)
                                                                  end
                                                                end
      allow(obj).to receive(:commit_atomizers!)
    end
  end

  def docker
    instance_double('Dapp::Docker').tap do |obj|
      allow(obj).to receive(:build_image!)
      allow(obj).to receive(:image_exist?).and_return(true)
    end
  end

  def config
    @config ||= {
        name: 'test',
        type: :shell,
        home_path: repo_path.join('.dapps/world'),
        from: :'ubuntu:16.04',
        git_artifact: {
            local: {
                where_to_add: '/app',
                cwd: '/',
                paths: nil,
                owner: 'apache',
                group: 'apache',
                interlayer_period: 604800
            }
        }
    }
  end

  def opts
    {
        log_indent: 0,
        dir: repo_path.join('.dapps'),
        build_path: repo_path.join('.dapps/world/build')
    }
  end

  # stages

  def stages
    @stages ||= build.stages.keys
  end

  def modifiable_stages
    [:prepare, :infra_install, :app_install, :infra_setup, :app_setup, :source_4, :source_5]
  end

  def next_modifiable_stages(stage)
    modifiable_stages[modifiable_stages.index(stage)+1..-1]
  end

  def build_and_check(stage)
    saved_keys = build_keys
    yield
    build_run
    send("#{stage}_saved").each { |s| expect(saved_keys).to include s => stage_build_key(s) }
    send("#{stage}_modified").each { |s| expect(saved_keys).to_not include s => stage_build_key(s) }
  end

  def build_keys
    build.stages.values.map { |s| [:"#{s.name}", s.signature] }.to_h
  end

  def stage_build_key(stage)
    build_keys[stage]
  end

  [:prepare, :infra_install, :app_install, :infra_setup, :app_setup, :source_4, :source_5].each do |stage|
    define_method stage do
      build_and_check(stage) { send(:"do_#{stage}") }
      next_modifiable_stages(stage).reverse_each { |s| send(s) }
    end

    define_method "#{stage}_modified" do
      stages[stages.index(stage)-1..-1]
    end

    define_method "#{stage}_saved" do
      stages[0..stages.index(stage)-2]
    end
  end

  [:infra_install, :app_install, :infra_setup, :app_setup].each do |stage|
    define_method :"do_#{stage}" do
      config[stage] = generate_command
    end
  end

  def do_prepare
    config[:from] = 'ubuntu:14.04'
  end

  def prepare_modified
    stages
  end

  def prepare_saved
    []
  end

  def infra_install_modified
    stages[stages.index(:infra_install)..-1]
  end

  def infra_install_saved
    [stages.first]
  end

  def do_source_4
    # change_file_and_commit
    # TODO
  end

  def source_4_saved # TODO
    # default
    stages
  end

  def source_4_modified # TODO
    # default
    []
  end

  def do_source_5
    change_file_and_commit
  end

  def source_5_saved # TODO
    # default
    stages[0..-2]
  end

  def source_5_modified # TODO
    # default
    [:source_5]
  end

  def change_file_and_commit(file='test_file', body=SecureRandom.hex)
    File.write repo_path.join(file), body
    commit!
  end

  def generate_command
    "echo '#{SecureRandom.hex}'"
  end

  def next_stage(s)
    build.stages[s].next.name
  end

  def prev_stage(s)
    build.stages[s].prev.name
  end

  # git

  def init_repo
    FileUtils.rm_rf repo_path
    FileUtils.mkpath repo_path
    git 'init'
    change_file_and_commit('README.md', 'Hello')
    change_file_and_commit('.gitignore', '.dapps/world/build')
    FileUtils.mkdir_p repo_path.join('.dapps/world')
    commit!
  end

  def repo_path
    Pathname('/tmp/dapp/hello')
  end

  def commit!
    git 'add --all'
    unless git('diff --cached --quiet', returns: [0, 1]).status.success?
      git 'commit -m +'
    end
  end

  def git(command, **kwargs)
    shellout "git -C #{repo_path} #{command}", **kwargs
  end
end
