require_relative 'spec_helper'

describe Dapp::Build::Shell do
  def current_build
    @build || build
  end

  def build_run
    build.run
  end

  def build
    options = { builder: builder, conf: config.dup, opts: opts }
    @build = Dapp::Build::Shell.new(**options).tap { |build| build.docker = docker }
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
    @docker ||= instance_double('Dapp::Docker').tap do |obj|
      allow(obj).to receive(:build_image!) { |image:, name:| images_cash << name }
      allow(obj).to receive(:image_exist?) { |name| images_cash.include? name }
    end
  end

  def images_cash
    @images_cash ||= []
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

  def stages
    @stages ||= build.stages.keys
  end

  def build_keys
    build.stages.values.map { |s| [:"#{s.name}", s.signature] }.to_h
  end

  def stage_build_key(stage)
    build_keys[stage]
  end

  def next_stage(s)
    build.stages[s].next.name
  end

  def prev_stage(s)
    build.stages[s].prev.name
  end

  def modifiable_stages
    [:prepare, :infra_install, :app_install, :infra_setup, :app_setup, :source_4, :source_5]
  end

  def next_modifiable_stages(stage)
    modifiable_stages[modifiable_stages.index(stage)+1..-1]
  end

  [:prepare, :infra_install, :app_install, :infra_setup, :app_setup, :source_4, :source_5].each do |stage|
    define_method stage do
      build_and_check(stage) { send(:"do_#{stage}") }
      next_modifiable_stages(stage).reverse_each { |s| puts "* #{s}"; send(s) }
    end

    define_method "#{stage}_modified" do
      stages[stages.index(stage)-1..-1]
    end

    define_method "#{stage}_saved" do
      stages[0..stages.index(stage)-2]
    end
  end

  [:source_4, :source_5].each do |stage|
    define_method "#{stage}_expectation" do
      check_image_command(stage, /git .* apply/)
    end
  end

  [:infra_install, :app_install, :infra_setup, :app_setup].each do |stage|
    define_method :"do_#{stage}" do
      config[stage] = generate_command
    end

    define_method "#{stage}_expectation" do
      check_image_command(stage, config[stage])
    end
  end

  def generate_command
    "echo '#{SecureRandom.hex}'"
  end

  def do_prepare
    config[:from] = 'ubuntu:14.04'
  end

  def prepare_expectation
    check_image_command(:prepare, 'apt-get update')
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
    file_path = repo_path.join('large_file')
    if File.exist? file_path
      FileUtils.rm file_path
      commit!
    else
      change_file_and_commit('large_file', ?x*Dapp::GitArtifact::MAX_PATCH_SIZE)
    end
  end

  def source_4_modified
    stages[stages.index(:source_4)..-1]
  end

  def do_source_5
    change_file_and_commit
  end

  def change_file_and_commit(file='test_file', body=SecureRandom.hex)
    File.write repo_path.join(file), body
    commit!
  end

  def source_5_saved
    stages[0..-2]
  end

  def source_5_modified
    [:source_5]
  end

  def build_and_check(stage)
    saved_keys = build_keys
    yield
    new_keys = build_keys
    build_run

    # caching
    modified, saved = current_build.stages.values.partition { |s| send("#{stage}_modified").include? s.name }
    modified.each { |s| expect(docker).to have_received(:build_image!).with(image: s.image, name: s.image_name) }
    saved.each { |s| expect(docker).to_not have_received(:build_image!).with(image: s.image, name: s.image_name) }

    # bash commands
    send("#{stage}_expectation")

    # signature
    send("#{stage}_saved").each { |s| expect(saved_keys).to include s => new_keys[s] }
    send("#{stage}_modified").each { |s| expect(saved_keys).to_not include s => new_keys[s] }
  end

  def check_image_command(stage, command)
    expect(current_build.stages[stage].image.build_cmd.join =~ Regexp.new(command)).to be
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

  it 'everything in the one right place' do
    init_repo
    build_run
    modifiable_stages.reverse_each { |s| puts s; send(s) }
  end
end
