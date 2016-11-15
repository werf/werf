require_relative '../spec_helper'

describe Dapp::Builder::Chef do
  include SpecHelper::Common
  include SpecHelper::Dimg

  before :all do
    init_project
  end

  %w(ubuntu:14.04 centos:7).each do |os|
    context os do
      it 'builds project' do
        [dimg, artifact_dimg].each do |d|
          %i(before_install install before_setup setup build_artifact).each do |stage|
            d.config._chef.send("__#{stage}_attributes")['mdapp-testartifact']['target_filename'] = 'CUSTOM_NAME_FROM_CHEF_SPEC.txt'
          end
        end

        dimg_build!

        stages.each { |_, stage| expect(stage.image.tagged?).to be(true) }

        TEST_FILE_NAMES.each { |name| expect(send("#{name}_exist?")).to be(true), "#{send("#{name}_path")} does not exist" }

        expect(file_exist_in_image?('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)).to be(true), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in artifact image'
        expect(file_exist_in_image?('/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(true), '/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in result image'

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)
        ).to eq(
          read_file_in_image('/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)
        ), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt in artifact image does not equal /myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt in result image'
      end

      [%i(before_install foo pizza batareika),
       %i(install bar taco koromyslo),
       %i(before_setup baz burger kolokolchik),
       %i(setup qux pelmeni taburetka)].each do |stage, project_file, mdapp_test_file, mdapp_test2_file|
        it "rebuilds from stage #{stage}" do
          old_template_file_values = {}
          old_template_file_values[project_file] = send(project_file)
          old_template_file_values[mdapp_test_file] = send(mdapp_test_file)
          old_template_file_values[mdapp_test2_file] = send(mdapp_test2_file)

          new_file_values = {}

          new_file_values[project_file] = SecureRandom.uuid
          testproject_chef_path.join("files/#{stage}/common/#{project_file}.txt").tap do |path|
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

          dimg_rebuild!

          expect(send(project_file, reload: true)).not_to eq(old_template_file_values[project_file])
          expect(send(mdapp_test_file, reload: true)).not_to eq(old_template_file_values[mdapp_test_file])
          expect(send(mdapp_test2_file, reload: true)).not_to eq(old_template_file_values[mdapp_test2_file])

          expect(send("test_#{stage}", reload: true)).to eq(new_file_values[project_file])
          expect(send("mdapp_test_#{stage}", reload: true)).to eq(new_file_values[mdapp_test_file])
          expect(send("mdapp_test2_#{stage}", reload: true)).to eq(new_file_values[mdapp_test2_file])
        end
      end

      xit 'rebuilds artifact from build_artifact stage' do
        old_artifact_before_install_stage_id = artifact_stages[:before_install].image.id
        old_artifact_last_stage_id = artifact_dimg.send(:last_stage).image.id

        [dimg, artifact_dimg].each do |d|
          %i(before_install install before_setup setup build_artifact).each do |stage|
            d.config._chef.send("__#{stage}_attributes")['mdapp-testartifact']['target_filename'] = 'SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt'
          end
        end

        dimg_rebuild!

        expect(artifact_stages[:before_install].image.id).to eq(old_artifact_before_install_stage_id)
        expect(artifact_dimg.send(:last_stage).image.id).not_to eq(old_artifact_last_stage_id)

        expect(file_exist_in_image?('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)).to be(true), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in artifact image'
        expect(file_exist_in_image?('/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(false), '/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does exist in result image'
        expect(file_exist_in_image?('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(true), '/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in result image'

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)
        ).to eq(
          read_file_in_image('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)
        ), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt inc artifact image does not equal /myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt in result image'
      end

      xit 'rebuilds artifact from before_install stage' do
        new_note_content = SecureRandom.uuid
        mdapp_testartifact_path.join('files/before_install/CUSTOM_NAME_FROM_CHEF_SPEC.txt').tap do |path|
          path.write "#{new_note_content}\n"
        end

        [dimg, artifact_dimg].each do |d|
          %i(before_install install before_setup setup build_artifact).each do |stage|
            d.config._chef.send("__#{stage}_attributes")['mdapp-testartifact']['target_filename'] = 'SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt'
          end
        end

        old_artifact_before_install_stage_id = artifact_stages[:before_install].image.id
        old_artifact_last_stage_id = artifact_dimg.send(:last_stage).image.id

        dimg_rebuild!

        expect(artifact_stages[:before_install].image.id).not_to eq(old_artifact_before_install_stage_id)
        expect(artifact_dimg.send(:last_stage).image.id).not_to eq(old_artifact_last_stage_id)

        expect(file_exist_in_image?('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)).to be(true), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in artifact image'
        expect(file_exist_in_image?('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(true), '/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in result image'

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)
        ).to eq(new_note_content)

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.send(:last_stage).image.name)
        ).to eq(
          read_file_in_image('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)
        ), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt inc artifact image does not equal /myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt in result image'
      end

      define_method :config do
        @config ||= default_config.merge(
          _builder: :chef,
          _name: "#{testproject_path.basename}-X-Y",
          _docker: default_config[:_docker].merge(_from: os.to_sym),
          _chef: {
            _dimod: %w(test test2),
            _recipe: %w(main X X_Y),
            __before_install_attributes: {
              'mdapp-test2' => {
                'sayhello' => 'hello',
                'sayhelloagain' => 'helloagain'
              },
              'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __install_attributes: {
              'mdapp-test2' => { 'sayhello' => 'hello' },
              'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __before_setup_attributes: {
              'mdapp-test2' => { 'sayhello' => 'hello' },
              'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __setup_attributes: {
              'mdapp-test2' => { 'sayhello' => 'hello' },
              'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __build_artifact_attributes: {
              'mdapp-test2' => { 'sayhello' => 'hello' },
              'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            }
          },
          _before_install_artifact: [
            ConfigRecursiveOpenStruct.new(
              _config: ConfigRecursiveOpenStruct.new(default_config.merge(
                _builder: :chef,
                _artifact_dependencies: [],
                _docker: default_config[:_docker].merge(_from: :'ubuntu:14.04'),
                _chef: {
                  _dimod: %w(testartifact),
                  _recipe: %w(myartifact),
                  __before_install_attributes: {
                    'mdapp-test2' => {
                      'sayhello' => 'hello',
                      'sayhelloagain' => 'helloagain'
                    },
                    'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __install_attributes: {
                    'mdapp-test2' => { 'sayhello' => 'hello' },
                    'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __before_setup_attributes: {
                    'mdapp-test2' => { 'sayhello' => 'hello' },
                    'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __setup_attributes: {
                    'mdapp-test2' => { 'sayhello' => 'hello' },
                    'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __build_artifact_attributes: {
                    'mdapp-test2' => { 'sayhello' => 'hello' },
                    'mdapp-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  }
                }
              )),
              _artifact_options: {
                cwd: '/',
                to: '/myartifact',
                exclude_paths: [],
                include_paths: []
              }
            )
          ]
        )
      end
    end # context
  end # each

  def openstruct_config
    ConfigRecursiveOpenStruct.new(config)
  end

  def project_path
    @project_path ||= Pathname("/tmp/dapp-test-#{SpecHelper::Dimg::CACHE_VERSION}")
  end

  def testproject_path
    project_path
  end

  def testproject_chef_path
    project_path.join('.dapp_chef')
  end

  def mdapp_test_path
    project_path.join('mdapp-test')
  end

  def mdapp_test2_path
    project_path.join('mdapp-test2')
  end

  def mdapp_testartifact_path
    project_path.join('mdapp-testartifact')
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

  def template_mdapp_testartifact_path
    @template_mdapp_testartifact_path ||= Pathname('spec/chef/mdapp-testartifact')
  end

  def init_project
    FileUtils.cp_r template_testproject_path, testproject_path.tap { |p| p.parent.mkpath }
    testproject_path.join('.dapps-build').tap { |p| p.rmtree if p.exist? }

    FileUtils.cp_r template_mdapp_test_path, mdapp_test_path.tap { |p| p.parent.mkpath }
    FileUtils.cp_r template_mdapp_test2_path, mdapp_test2_path.tap { |p| p.parent.mkpath }
    FileUtils.cp_r template_mdapp_testartifact_path, mdapp_testartifact_path.tap { |p| p.parent.mkpath }
  end
  # rubocop:enable Metrics/AbcSize

  def artifact_dimg
    stages[:before_install_artifact].send(:artifacts).first[:dimg]
  end

  def artifact_stages
    _stages_of_dimg(artifact_dimg)
  end

  TEST_FILE_NAMES = %i(foo X_foo X_Y_foo bar baz qux
                       burger pizza taco pelmeni
                       kolokolchik koromyslo taburetka batareika
                       test_before_install test_install
                       test_before_setup test_setup
                       mdapp_test_before_install mdapp_test_install
                       mdapp_test_before_setup mdapp_test_setup
                       mdapp_test2_before_install mdapp_test2_install
                       mdapp_test2_before_setup mdapp_test2_setup).freeze

  TEST_FILE_NAMES.each do |name|
    define_method("#{name}_path") do
      "/#{name}.txt"
    end

    define_method(name) do |reload: false|
      (!reload && instance_variable_get("@#{name}")) ||
        instance_variable_set("@#{name}", read_file_in_image(send("#{name}_path"), dimg.send(:last_stage).image.name))
    end

    define_method("#{name}_exist?") do
      file_exist_in_image? send("#{name}_path"), dimg.send(:last_stage).image.name
    end
  end

  def read_file_in_image(path, image_name)
    shellout!("docker run --rm #{image_name} cat #{path}").stdout.strip
  end

  def file_exist_in_image?(path, image_name)
    res = shellout("docker run --rm #{image_name} ls #{path}")
    return true if res.exitstatus.zero?
    return false if res.exitstatus == 2
    res.error!
  end

  class ConfigRecursiveOpenStruct < RecursiveOpenStruct
    def _dimg_chain
      [self]
    end

    def _dimg_runlist
      []
    end

    def _root_dimg
      _dimg_chain.first
    end

    def to_json(*a)
      to_h.to_json(*a)
    end
  end
end
