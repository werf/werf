require_relative '../spec_helper'

describe Dapp::Application do
  include SpecHelpers::Common
  include SpecHelpers::Application
  include SpecHelpers::Git

  before :all do
    @wd = Dir.pwd
    init
  end

  before :each do
    stub_docker_image
    application_build!
  end

  after :all do
    Dir.chdir @wd
  end

  def init
    FileUtils.rm_rf project_path
    FileUtils.mkpath project_path
    Dir.chdir project_path
    git_init!
  end

  def project_path
    Pathname('/tmp/dapp/test')
  end

  def project_dapp_path
    project_path.join('.dapps/dapp')
  end

  def config
    @config ||= RecursiveOpenStruct.new(
      _name: 'test', _builder: :shell, _home_path: project_path,
      _shell: { _infra_install: ['apt-get update', 'apt-get -y dist-upgrade',
                                 'apt-get -y install apt-utils curl apt-transport-https git'],
                _infra_setup: [], _app_install: [], _app_setup: [] },
      _docker: { _from: :'ubuntu:16.04', _expose: [] },
      _git_artifact: { _local: { _artifact_options: { where_to_add: '/app' } } }
    )
  end

  def cli_options
    @cli_options ||= default_cli_options.merge(build_dir: project_dapp_path.join('build'))
  end

  [:from, :infra_install, :app_install, :infra_setup, :app_setup, :source_4, :source_5].each do |stage_name|
    define_method "#{stage_name}_modified_signatures" do
      stages_names[stages_names.index(stage_name) - 1..-1]
    end

    define_method "#{stage_name}_saved_signatures" do
      stages_names[0..stages_names.index(stage_name) - 2]
    end
  end

  [:infra_install, :app_install, :infra_setup, :app_setup].each do |stage_name|
    define_method :"change_#{stage_name}" do
      config._shell.send("_#{stage_name}") << generate_command
    end
  end

  [:app_install, :infra_setup, :app_setup].each do |stage_name|
    define_method "expect_#{stage_name}_image" do
      check_image_command(stage_name, config._shell.send("_#{stage_name}").last)
      check_image_command(prev_stage(stage_name), 'apply')
    end
  end

  [:source_4, :source_5].each do |stage_name|
    define_method "expect_#{stage_name}_image" do
      check_image_command(stage_name, 'apply')
    end
  end

  def check_image_command(stage_name, command)
    expect(stages[stage_name].send(:image).send(:bash_commands).join =~ Regexp.new(command)).to be
  end

  def change_from
    config._docker._from = 'ubuntu:14.04'
  end

  def expect_from_image
    check_image_command(:infra_install, 'update')
    check_image_command(:source_1_archive, 'tar -x')
  end

  def from_modified_signatures
    stages_names
  end

  def from_saved_signatures
    []
  end

  def infra_install_modified_signatures
    stages_names[stages_names.index(:infra_install)..-1]
  end

  def infra_install_saved_signatures
    [stages_names.first]
  end

  def expect_infra_install_image
    check_image_command(:infra_install, config._shell._infra_install.last)
    check_image_command(:source_1_archive, 'tar -x')
  end

  def change_source_4
    file_path = project_path.join('large_file')
    if File.exist? file_path
      FileUtils.rm file_path
      git_commit!
    else
      git_change_and_commit!('large_file', 'x' * 1024 * 1024)
    end
  end

  def source_4_modified_signatures
    stages_names[stages_names.index(:source_4)..-1]
  end

  def change_source_5
    git_change_and_commit!
  end

  def source_5_saved_signatures
    stages_names[0..-2]
  end

  def source_5_modified_signatures
    [:source_5]
  end

  def build_and_check(stage_name)
    check_signatures_and_build(stage_name)
    expect_built_stages(stage_name)
    send("expect_#{stage_name}_image")
  end

  def check_signatures_and_build(stage_name)
    saved_signatures = stages_signatures
    send(:"change_#{stage_name}")
    application_renew
    expect_stages_signatures(stage_name, saved_signatures, stages_signatures)
    application_build!
  end

  def expect_built_stages(stage_name)
    parted_stages_signatures(stage_name) do |built_stages, not_built_stages|
      not_built_stages.each do |s|
        expect(s.send(:image)).to_not have_received(:build!)
        expect(s.send(:image)).to_not have_received(:tag!)
      end
      built_stages.each do |s|
        expect(s.send(:image)).to have_received(:build!)
        expect(s.send(:image)).to have_received(:tag!)
      end
    end
  end

  def parted_stages_signatures(stage_name)
    built_stages, not_built_stages = stages.values.partition do |s|
      send("#{stage_name}_modified_signatures").include? s.send(:name)
    end
    yield built_stages, not_built_stages
  end

  def expect_stages_signatures(stage_name, saved_keys, new_keys)
    send("#{stage_name}_saved_signatures").each { |s| expect(saved_keys).to include s => new_keys[s] }
    send("#{stage_name}_modified_signatures").each { |s| expect(saved_keys).to_not include s => new_keys[s] }
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

  def from
    build_and_check(:from)
    source_5
    source_4
    app_setup
    infra_setup
    app_install
    infra_install
  end

  [:source_5, :source_4, :app_setup, :infra_setup, :app_install, :infra_install, :from].each do |stage|
    it "test #{stage}" do
      send(stage)
    end
  end
end
