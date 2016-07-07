require_relative '../spec_helper'

describe Dapp::Builder::Chef do
  include SpecHelpers::Common
  include SpecHelpers::Application

  before :all do
    init_project
  end

  it "builds chef testproject" do
    application_build!
    stages.each {|_, stage| expect(stage.image.exist?).to be(true)}
    TEST_FILE_NAMES.each {|name| expect(send("#{name}_exist?")).to be(true)}
  end

  %i(infra_install app_install infra_setup app_setup).each do |stage|
    it "changes chef stage #{stage}" do
      #TODO
    end
  end

  def config
    @config ||= {
      name: 'testproject',
      type: :chef,
      from: 'ubuntu:14.04',
      home_path: testproject_path.to_s,
    }
  end

  def project_path
    @project_path ||= Pathname("/tmp/dapp-test-#{SecureRandom.uuid}")
  end

  def testproject_path
    project_path.join('testproject')
  end

  def mdapp_test_path
    project_path.join('mdapp-test')
  end

  def template_testproject_path
    @template_testproject_path ||= Pathname('spec/chef/testproject')
  end

  def template_mdapp_test_path
    @template_mdapp_test_path ||= Pathname('spec/chef/mdapp-test')
  end

  def init_project
    FileUtils.cp_r template_testproject_path, testproject_path.tap {|p| p.parent.mkpath}
    FileUtils.cp_r template_mdapp_test_path, mdapp_test_path.tap {|p| p.parent.mkpath}
  end

  TEST_FILE_NAMES = %i(foo bar baz qux burger pizza taco pelmeni
                       testproject_infra_install testproject_app_install
                       testproject_infra_setup testproject_app_setup
                       mdapp_test_infra_install mdapp_test_app_install
                       mdapp_test_infra_setup mdapp_test_app_setup)

  TEST_FILE_NAMES.each do |name|
    define_method(name) do |reload: false|
      (!reload && instance_variable_get("@#{name}")) ||
        instance_variable_set("@#{name}",
          shellout!("docker run --rm #{application.last_stage.image.name} cat /#{name}.txt").stdout.strip)
    end

    define_method("#{name}_exist?") do
      res = shellout("docker run --rm #{application.last_stage.image.name} ls /#{name}.txt")
      return true if res.exitstatus == 0
      return false if res.exitstatus == 2
      res.error!
    end
  end
end
