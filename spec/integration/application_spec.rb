require_relative '../spec_helper'

describe Dapp::Application do
  def current_application
    @application || application
  end

  def application_build!
    application.build_and_fixate!
  end

  def application
    options = { conf: config.dup, opts: opts }
    @application = Dapp::Application.new(**options)
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
    @opts ||= {
        log_indent: 0,
        dir: repo_path.join('.dapps'),
        build_path: repo_path.join('.dapps/world/build')
    }
  end

  def stages_names
    @stages ||= stages.keys.reverse
  end

  def stages(b=application)
    stgs = {}
    s = b.last_stage
    while s.respond_to? :prev_stage
      stgs[s.send(:name)] = s
      s = s.prev_stage
    end
    stgs
  end

  def build_keys
    stages.values.map { |s| [:"#{s.send(:name)}", s.send(:signature)] }.to_h
  end

  def stage_build_key(stage_name)
    build_keys[stage_name]
  end

  def next_stage(s)
    stages(current_application)[s].next_stage.send(:name)
  end

  def prev_stage(s)
    stages(current_application)[s].prev_stage.send(:name)
  end


  [:prepare, :infra_install, :app_install, :infra_setup, :app_setup, :source_4, :source_5].each do |stage_name|
    define_method "#{stage_name}_modified_signatures" do
      stages_names[stages_names.index(stage_name)-1..-1]
    end

    define_method "#{stage_name}_saved_signatures" do
      stages_names[0..stages_names.index(stage_name)-2]
    end
  end

  [:infra_install, :app_install, :infra_setup, :app_setup].each do |stage_name|
    define_method :"change_#{stage_name}" do
      config[stage_name] = generate_command
    end
  end

  [:app_install, :infra_setup, :app_setup].each do |stage_name|
    define_method "expect_#{stage_name}_images_commands" do
      check_image_command(stage_name, config[stage_name])
      check_image_command(prev_stage(stage_name), 'patch')
    end
  end

  [:source_4, :source_5].each do |stage_name|
    define_method "expect_#{stage_name}_images_commands" do
      check_image_command(stage_name, 'patch')
    end
  end

  def check_image_command(stage_name, command)
    expect(stages(current_application)[stage_name].send(:image).bash_commands.join =~ Regexp.new(command)).to be
  end

  def generate_command
    "echo '#{SecureRandom.hex}'"
  end

  def change_prepare
    config[:from] = 'ubuntu:14.04'
  end

  def expect_prepare_images_commands
    check_image_command(:prepare, 'apt-get update')
    check_image_command(:source_1_archive, 'tar xf')
  end

  def prepare_modified_signatures
    stages_names
  end

  def prepare_saved_signatures
    []
  end

  def infra_install_modified_signatures
    stages_names[stages_names.index(:infra_install)..-1]
  end

  def infra_install_saved_signatures
    [stages_names.first]
  end

  def expect_infra_install_images_commands
    check_image_command(:infra_install, config[:infra_install])
    check_image_command(:source_1_archive, 'tar xf')
  end

  def change_source_4
    file_path = repo_path.join('large_file')
    if File.exist? file_path
      FileUtils.rm file_path
      commit!
    else
      change_file_and_commit('large_file', ?x*1024*1024)
    end
  end

  def source_4_modified_signatures
    stages_names[stages_names.index(:source_4)..-1]
  end

  def change_source_5
    change_file_and_commit
  end

  def change_file_and_commit(file='test_file', body=SecureRandom.hex)
    File.write repo_path.join(file), "#{body}\n"
    commit!
  end

  def source_5_saved_signatures
    stages_names[0..-2]
  end

  def source_5_modified_signatures
    [:source_5]
  end


  def build_and_check(stage_name)
    check_signatures_and_build(stage_name)
    # expect_built_stages(stage_name) TODO
    send("expect_#{stage_name}_images_commands")
  end

  def check_signatures_and_build(stage_name)
    saved_signatures = build_keys
    send(:"change_#{stage_name}")
    expect_stages_signatures(stage_name, saved_signatures, build_keys)
    application_build!
  end

  def expect_built_stages(stage_name)
    built_stages = stages(current_application).values.select { |s| send("#{stage_name}_modified_signatures").include? s.send(:name) }
    built_stages.each { |s| expect(docker).to have_received(:build_image!).with(image_specification: s.send(:image), image_name: s.send(:image_name)) }
  end

  def expect_stages_signatures(stage_name, saved_keys, new_keys)
    send("#{stage_name}_saved_signatures").each { |s| expect(saved_keys).to include s => new_keys[s] }
    send("#{stage_name}_modified_signatures").each { |s| expect(saved_keys).to_not include s => new_keys[s] }
  end


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


  def source_5
    build_and_check(:source_5)
  end

  def source_4
    build_and_check(:source_4)
    source_5
  end

  def app_setup
    build_and_check(:app_setup)
    source_5
    source_4
  end

  def infra_setup
    build_and_check(:infra_setup)
    source_5
    source_4
    app_setup
  end

  def app_install
    build_and_check(:app_install)
    source_5
    source_4
    app_setup
    infra_setup
  end

  def infra_install
    build_and_check(:infra_install)
    source_5
    source_4
    app_setup
    infra_setup
    app_install
  end

  def prepare
    build_and_check(:prepare)
    source_5
    source_4
    app_setup
    infra_setup
    app_install
    infra_install
  end


  before :all do
    init_repo
    application_build!
  end

  [:source_5, :source_4, :app_setup, :infra_setup, :app_install, :infra_install, :prepare].each do |stage|
    it "test #{stage}" do
      send(stage)
    end
  end
end
