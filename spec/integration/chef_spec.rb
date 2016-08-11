require_relative '../spec_helper'

describe Dapp::Builder::Chef do
  include SpecHelper::Common
  include SpecHelper::Application

  CACHE_VERSION = SecureRandom.uuid

  before :all do
    init_project
  end

  %w(ubuntu:14.04 centos:7).each do |os|
    context os do
      it 'builds project' do
        application_build!
        stages.each { |_, stage| expect(stage.image.tagged?).to be(true) }
        TEST_FILE_NAMES.each { |name| expect(send("#{name}_exist?")).to be(true), "#{send("#{name}_path")} does not exist" }
      end

      [%i(infra_install foo pizza batareika),
       %i(install bar taco koromyslo),
       %i(infra_setup baz burger kolokolchik),
       %i(setup qux pelmeni taburetka)].each do |stage, project_file, mdapp_test_file, mdapp_test2_file|
        it "rebuilds from stage #{stage}" do
          old_template_file_values = {}
          old_template_file_values[project_file] = send(project_file)
          old_template_file_values[mdapp_test_file] = send(mdapp_test_file)
          old_template_file_values[mdapp_test2_file] = send(mdapp_test2_file)

          new_file_values = {}

          new_file_values[project_file] = SecureRandom.uuid
          testproject_path.join("files/#{stage}/#{project_file}.txt").tap do |path|
            path.write "#{new_file_values[project_file]}\n"
          end

          new_file_values[mdapp_test_file] = SecureRandom.uuid
          mdapp_test_path.join("files/#{stage}/#{mdapp_test_file}.txt").tap do |path|
            path.write "#{new_file_values[mdapp_test_file]}\n"
          end

          new_file_values[mdapp_test2_file] = SecureRandom.uuid
          mdapp_test2_path.join("files/#{stage}/#{mdapp_test2_file}.txt").tap do |path|
            path.write "#{new_file_values[mdapp_test2_file]}\n"
          end

          application_rebuild!

          expect(send(project_file, reload: true)).not_to eq(old_template_file_values[project_file])
          expect(send(mdapp_test_file, reload: true)).not_to eq(old_template_file_values[mdapp_test_file])
          expect(send(mdapp_test2_file, reload: true)).not_to eq(old_template_file_values[mdapp_test2_file])

          expect(send("test_#{stage}", reload: true)).to eq(new_file_values[project_file])
          expect(send("mdapp_test_#{stage}", reload: true)).to eq(new_file_values[mdapp_test_file])
          expect(send("mdapp_test2_#{stage}", reload: true)).to eq(new_file_values[mdapp_test2_file])
        end
      end

      define_method :config do
        @config ||= default_config.merge(
          _builder: :chef,
          _home_path: testproject_path.to_s,
          _name: "#{testproject_path.basename}-X-Y",
          _chef: {
            _modules: %w(test test2),
            _recipes: %w(main X X_Y)
          }
        ).tap do |config|
          config[:_docker][:_from] = os.to_sym
          config[:_docker][:_from_cache_version] = CACHE_VERSION
        end
      end
    end # context
  end # each

  def openstruct_config
    RecursiveOpenStruct.new(config).tap do |obj|
      def obj._app_chain
        [self]
      end

      def obj._app_runlist
        []
      end

      def obj._root_app
        _app_chain.first
      end
    end
  end

  def project_path
    @project_path ||= Pathname("/tmp/dapp-test-#{CACHE_VERSION}")
  end

  def testproject_path
    project_path.join('testproject')
  end

  def mdapp_test_path
    project_path.join('mdapp-test')
  end

  def mdapp_test2_path
    project_path.join('mdapp-test2')
  end

  def template_testproject_path
    @template_testproject_path ||= Pathname('spec/chef/testproject')
  end

  def template_mdapp_test_path
    @template_mdapp_test_path ||= Pathname('spec/chef/mdapp-test')
  end

  def template_mdapp_test2_path
    @template_mdapp_test2_path ||= Pathname('spec/chef/mdapp-test2')
  end

  def init_project
    FileUtils.cp_r template_testproject_path, testproject_path.tap { |p| p.parent.mkpath }
    testproject_path.join('.dapps-build').tap { |p| p.rmtree if p.exist? }

    FileUtils.cp_r template_mdapp_test_path, mdapp_test_path.tap { |p| p.parent.mkpath }
    FileUtils.cp_r template_mdapp_test2_path, mdapp_test2_path.tap { |p| p.parent.mkpath }
  end
  # rubocop:enable Metrics/AbcSize

  TEST_FILE_NAMES = %i(foo X_foo X_Y_foo bar baz qux
                       burger pizza taco pelmeni
                       kolokolchik koromyslo taburetka batareika
                       test_infra_install test_install
                       test_infra_setup test_setup
                       mdapp_test_infra_install mdapp_test_install
                       mdapp_test_infra_setup mdapp_test_setup
                       mdapp_test2_infra_install mdapp_test2_install
                       mdapp_test2_infra_setup mdapp_test2_setup).freeze

  TEST_FILE_NAMES.each do |name|
    define_method("#{name}_path") do
      "/#{name}.txt"
    end

    define_method(name) do |reload: false|
      (!reload && instance_variable_get("@#{name}")) ||
        instance_variable_set("@#{name}",
                              shellout!("docker run --rm #{application.send(:last_stage).image.name} cat #{send("#{name}_path")}").stdout.strip)
    end

    define_method("#{name}_exist?") do
      res = shellout("docker run --rm #{application.send(:last_stage).image.name} ls #{send("#{name}_path")}")
      return true if res.exitstatus.zero?
      return false if res.exitstatus == 2
      res.error!
    end
  end
end
