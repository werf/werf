require_relative '../spec_helper'

describe Dapp::Dimg::Builder::Chef do
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
            d.config._chef.send("__#{stage}_attributes")['dimod-testartifact']['target_filename'] = 'CUSTOM_NAME_FROM_CHEF_SPEC.txt'
          end
        end

        dimg_build!

        stages.each { |_, stage| expect(stage.image.tagged?).to be(true) }

        TEST_FILE_NAMES.each { |name| expect(send("#{name}_exist?")).to be(true), "#{send("#{name}_path")} does not exist" }

        expect(file_exist_in_image?('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)).to be(true), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in artifact image'
        expect(file_exist_in_image?('/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(true), '/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in result image'

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)
        ).to eq(
          read_file_in_image('/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)
        ), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt in artifact image does not equal /myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt in result image'
      end

      [%i(before_install foo pizza batareika),
       %i(install bar taco koromyslo),
       %i(before_setup baz burger kolokolchik),
       %i(setup qux pelmeni taburetka)].each do |stage, project_file, dimod_test_file, dimod_test2_file|
        it "rebuilds from stage #{stage}" do
          old_template_file_values = {}
          old_template_file_values[project_file] = send(project_file)
          old_template_file_values[dimod_test_file] = send(dimod_test_file)
          old_template_file_values[dimod_test2_file] = send(dimod_test2_file)

          new_file_values = {}

          new_file_values[project_file] = SecureRandom.uuid
          testproject_chef_path.join("files/#{stage}/common/#{project_file}.txt").tap do |path|
            path.write "#{new_file_values[project_file]}\n"
          end

          new_file_values[dimod_test_file] = SecureRandom.uuid
          dimod_test_path.join("files/#{stage}/#{dimod_test_file}.txt").tap do |path|
            path.write "#{new_file_values[dimod_test_file]}\n"
          end

          new_file_values[dimod_test2_file] = SecureRandom.uuid
          dimod_test2_path.join("files/#{stage}/#{dimod_test2_file}.txt").tap do |path|
            path.write "#{new_file_values[dimod_test2_file]}\n"
          end

          dimg_rebuild!

          expect(send(project_file, reload: true)).not_to eq(old_template_file_values[project_file])
          expect(send(dimod_test_file, reload: true)).not_to eq(old_template_file_values[dimod_test_file])
          expect(send(dimod_test2_file, reload: true)).not_to eq(old_template_file_values[dimod_test2_file])

          expect(send("dapp_#{stage}", reload: true)).to eq(new_file_values[project_file])
          expect(send("dimod_test_#{stage}", reload: true)).to eq(new_file_values[dimod_test_file])
          expect(send("dimod_test2_#{stage}", reload: true)).to eq(new_file_values[dimod_test2_file])
        end
      end

      xit 'rebuilds artifact from build_artifact stage' do
        old_artifact_before_install_stage_id = artifact_stages[:before_install].image.id
        old_artifact_last_stage_id = artifact_dimg.last_stage.image.id

        [dimg, artifact_dimg].each do |d|
          %i(before_install install before_setup setup build_artifact).each do |stage|
            d.config._chef.send("__#{stage}_attributes")['dimod-testartifact']['target_filename'] = 'SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt'
          end
        end

        dimg_rebuild!

        expect(artifact_stages[:before_install].image.id).to eq(old_artifact_before_install_stage_id)
        expect(artifact_dimg.last_stage.image.id).not_to eq(old_artifact_last_stage_id)

        expect(file_exist_in_image?('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)).to be(true), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in artifact image'
        expect(file_exist_in_image?('/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(false), '/myartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does exist in result image'
        expect(file_exist_in_image?('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(true), '/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in result image'

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)
        ).to eq(
          read_file_in_image('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)
        ), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt inc artifact image does not equal /myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt in result image'
      end

      xit 'rebuilds artifact from before_install stage' do
        new_note_content = SecureRandom.uuid
        dimod_testartifact_path.join('files/before_install/CUSTOM_NAME_FROM_CHEF_SPEC.txt').tap do |path|
          path.write "#{new_note_content}\n"
        end

        [dimg, artifact_dimg].each do |d|
          %i(before_install install before_setup setup build_artifact).each do |stage|
            d.config._chef.send("__#{stage}_attributes")['dimod-testartifact']['target_filename'] = 'SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt'
          end
        end

        old_artifact_before_install_stage_id = artifact_stages[:before_install].image.id
        old_artifact_last_stage_id = artifact_dimg.last_stage.image.id

        dimg_rebuild!

        expect(artifact_stages[:before_install].image.id).not_to eq(old_artifact_before_install_stage_id)
        expect(artifact_dimg.last_stage.image.id).not_to eq(old_artifact_last_stage_id)

        expect(file_exist_in_image?('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)).to be(true), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in artifact image'
        expect(file_exist_in_image?('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)).to be(true), '/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt does not exist in result image'

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)
        ).to eq(new_note_content)

        expect(
          read_file_in_image('/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt', artifact_dimg.last_stage.image.name)
        ).to eq(
          read_file_in_image('/myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt', dimg.send(:last_stage).image.name)
        ), '/testartifact/CUSTOM_NAME_FROM_CHEF_SPEC.txt inc artifact image does not equal /myartifact/SECOND_CUSTOM_NAME_FROM_CHEF_SPEC.txt in result image'
      end

      define_method :config do
        @config ||= default_config.merge(
          _builder: :chef,
          _name: "#{testproject_path.basename}-x-y",
          _docker: default_config[:_docker].merge(_from: os.to_sym),
          _chef: {
            _dimod: ['dimod-test', 'dimod-test2'],
            _recipe: %w(main x x_y),
            _cookbook: ConfigHash.new(
              'build-essential' => {name: 'build-essential', version_constraint: '~> 8.0.0'},
              'dimod-test' => {name: 'dimod-test', path: File.expand_path('../dimod-test', dapp.path)},
              'dimod-test2' => {name: 'dimod-test2', path: File.expand_path('../dimod-test2', dapp.path)}
            ),
            __before_install_attributes: {
              'dimod-test2' => {
                'sayhello' => 'hello',
                'sayhelloagain' => 'helloagain'
              },
              'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __install_attributes: {
              'dimod-test2' => { 'sayhello' => 'hello' },
              'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __before_setup_attributes: {
              'dimod-test2' => { 'sayhello' => 'hello' },
              'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __setup_attributes: {
              'dimod-test2' => { 'sayhello' => 'hello' },
              'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            },
            __build_artifact_attributes: {
              'dimod-test2' => { 'sayhello' => 'hello' },
              'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
            }
          },
          _before_install_artifact: [
            ConfigRecursiveOpenStruct.new(
              _config: ConfigRecursiveOpenStruct.new(default_config.merge(
                _builder: :chef,
                _docker: default_config[:_docker].merge(_from: :'ubuntu:14.04'),
                _chef: {
                  _dimod: ['dimod-testartifact'],
                  _recipe: %w(myartifact),
                  _cookbook: ConfigHash.new(
                    'build-essential' => {name: 'build-essential', version_constraint: '~> 8.0.0'},
                    'dimod-testartifact' => {name: 'dimod-testartifact', path: File.expand_path('../dimod-testartifact', dapp.path)}
                  ),
                  __before_install_attributes: {
                    'dimod-test2' => {
                      'sayhello' => 'hello',
                      'sayhelloagain' => 'helloagain'
                    },
                    'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __install_attributes: {
                    'dimod-test2' => { 'sayhello' => 'hello' },
                    'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __before_setup_attributes: {
                    'dimod-test2' => { 'sayhello' => 'hello' },
                    'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __setup_attributes: {
                    'dimod-test2' => { 'sayhello' => 'hello' },
                    'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  },
                  __build_artifact_attributes: {
                    'dimod-test2' => { 'sayhello' => 'hello' },
                    'dimod-testartifact' => { 'target_filename' => 'CUSTOM_NAME_FROM_CHEF_SPEC.txt' }
                  }
                }
              )),
              _artifact_options: {
                cwd: '/myartifact_testproject',
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

  def _base_path
    @_base_path ||= Pathname("/tmp/dapp-test-#{SpecHelper::Dimg::CACHE_VERSION}")
  end

  def project_path
    testproject_path
  end

  def testproject_path
    _base_path.join('testproject')
  end

  def testproject_chef_path
    testproject_path.join('.dapp_chef')
  end

  def dimod_test_path
    _base_path.join('dimod-test')
  end

  def dimod_test2_path
    _base_path.join('dimod-test2')
  end

  def dimod_testartifact_path
    _base_path.join('dimod-testartifact')
  end

  def template_testproject_path
    @template_testproject_path ||= Pathname('spec/chef/testproject')
  end

  def template_dimod_test_path
    @template_dimod_test_path ||= Pathname('spec/chef/dimod-test')
  end

  def template_dimod_test2_path
    @template_dimod_test2_path ||= Pathname('spec/chef/dimod-test2')
  end

  def template_dimod_testartifact_path
    @template_dimod_testartifact_path ||= Pathname('spec/chef/dimod-testartifact')
  end

  def init_project
    FileUtils.cp_r template_testproject_path, testproject_path.tap { |p| p.parent.mkpath }
    testproject_path.join('.dapp_build').tap { |p| p.rmtree if p.exist? }

    FileUtils.cp_r template_dimod_test_path, dimod_test_path.tap { |p| p.parent.mkpath }
    FileUtils.cp_r template_dimod_test2_path, dimod_test2_path.tap { |p| p.parent.mkpath }
    FileUtils.cp_r template_dimod_testartifact_path, dimod_testartifact_path.tap { |p| p.parent.mkpath }
  end
  # rubocop:enable Metrics/AbcSize

  def artifact_dimg
    stages[:before_install_artifact].send(:artifacts).first[:dimg]
  end

  def artifact_stages
    _stages_of_dimg(artifact_dimg)
  end

  TEST_FILE_NAMES = %i(foo x_foo x_y_foo bar baz qux
                       burger pizza taco pelmeni
                       kolokolchik koromyslo taburetka batareika
                       dapp_before_install dapp_install
                       dapp_before_setup dapp_setup
                       dimod_test_before_install dimod_test_install
                       dimod_test_before_setup dimod_test_setup
                       dimod_test2_before_install dimod_test2_install
                       dimod_test2_before_setup dimod_test2_setup).freeze

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

  class ConfigHash
    def initialize(hash)
      @data = hash
    end

    def method_missing(*args, &blk)
      @data.send(*args, &blk)
    end
  end
end
