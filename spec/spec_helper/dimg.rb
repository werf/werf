module SpecHelper
  module Dimg
    CACHE_VERSION = SecureRandom.uuid

    def dimg_build!
      dimg.build!
    end

    def dimg
      @dimg || dimg_renew
    end

    def dimg_renew
      @openstruct_config = nil
      @dimg = begin
        options = { config: openstruct_config, dapp: dapp }
        Dapp::Dimg::Dimg.new(**options)
      end
    end

    def dimg_rebuild!
      dimg_renew
      dimg_build!
    end

    def dapp
      @dapp ||= begin
        allow_any_instance_of(Dapp::Dapp).to receive(:dappfile_path) { File.join(project_path, 'Dappfile') }
        allow_any_instance_of(Dapp::Dapp).to receive(:config) { config }
        yield if block_given?
        Dapp::Dapp.new(options: dapp_options)
      end
    end

    def project_path
      @project_path ||= Dir.mktmpdir
    end

    def openstruct_config
      @openstruct_config ||= RecursiveOpenStruct.new(config)
    end

    def config
      @config ||= default_config
    end

    def default_config
      Marshal.load(Marshal.dump(_name: 'test',
                                      _import_artifact: [],
                                      _before_install_artifact: [], _before_setup_artifact: [],
                                      _after_install_artifact: [], _after_setup_artifact: [],
                                      _tmp_dir_mount: [], _build_dir_mount: [], _custom_dir_mount: [],
                                      _chef: { _dimod: [], _recipe: [] },
                                      _shell: { _before_install_command: [], _before_setup_command: [],
                                                _install_command: [], _setup_command: [] },
                                      _docker: { _from: :'ubuntu:14.04',
                                                 _from_cache_version: CACHE_VERSION,
                                                 _change_options: {} },
                                      _git_artifact: { _local: [], _remote: [] }))
    end

    def dapp_options
      default_dapp_options
    end

    def default_dapp_options
      { quiet: true, color: 'off', dev: false }
    end

    def stages
      _stages_of_dimg(dimg)
    end

    def _stages_of_dimg(dimg)
      hash = {}
      s = dimg.send(:last_stage)
      while s.respond_to? :prev_stage
        hash[s.send(:name)] = s
        s = s.prev_stage
      end
      hash
    end

    def stage_signature(stage_name)
      stages_signatures[stage_name]
    end

    def next_stage(s)
      stages[s].next_stage.send(:name)
    end

    def prev_stage(s)
      stages[s].prev_stage.send(:name)
    end

    def stub_dimg
      method_new = Dapp::Dimg::Dimg.method(:new)

      dimg = class_double(Dapp::Dimg::Dimg).as_stubbed_const
      allow(dimg).to receive(:new) do |*args, &block|
        method_new.call(*args, &block).tap do |instance|
          allow(instance).to receive(:home_path) { |*m_args| Pathname(File.absolute_path(File.join(*m_args))) }
          allow(instance).to receive(:filelock)
          yield instance if block_given?
        end
      end
    end

    def empty_dimg
      Dapp::Dimg::Dimg.new(dapp: nil, config: openstruct_config)
    end

    def empty_artifact
      Dapp::Dimg::Artifact.new(dapp: nil, config: openstruct_config)
    end
  end
end
