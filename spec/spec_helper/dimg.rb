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
        options = { config: openstruct_config, project: project }
        Dapp::Dimg.new(**options)
      end
    end

    def dimg_rebuild!
      dimg_renew
      dimg_build!
    end

    def project
      @project ||= begin
        allow_any_instance_of(Dapp::Project).to receive(:dappfiles) { [File.join(project_path || Dir.mktmpdir, 'Dappfile')] }
        yield if block_given?
        Dapp::Project.new(cli_options: cli_options)
      end
    end

    def project_path
    end

    def openstruct_config
      @openstruct_config ||= RecursiveOpenStruct.new(config)
    end

    def config
      @config ||= default_config
    end

    def default_config
      Marshal.load(Marshal.dump(_basename: 'dapp',
                                _name: 'test',
                                _import_artifact: [],
                                _before_install_artifact: [], _before_setup_artifact: [],
                                _after_install_artifact: [], _after_setup_artifact: [],
                                _tmp_dir: { _store: [] }, _build_dir: { _store: [] },
                                _chef: { _modules: [] },
                                _shell: { _before_install: [], _before_setup: [], _install: [], _setup: [] },
                                _docker: { _from: :'ubuntu:14.04',
                                           _from_cache_version: CACHE_VERSION,
                                           _change_options: {} },
                                _git_artifact: { _local: [], _remote: [] },
                                _install_dependencies: [], _setup_dependencies: []))
    end

    def cli_options
      default_cli_options
    end

    def default_cli_options
      { log_quiet: true, log_color: 'off' }
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
      method_new = Dapp::Dimg.method(:new)

      dimg = class_double(Dapp::Dimg).as_stubbed_const
      allow(dimg).to receive(:new) do |*args, &block|
        method_new.call(*args, &block).tap do |instance|
          allow(instance).to receive(:home_path) { |*m_args| Pathname(File.absolute_path(File.join(*m_args))) }
          allow(instance).to receive(:filelock)
        end
      end
    end

    def empty_dimg
      Dapp::Dimg.new(project: nil, config: openstruct_config)
    end

    def empty_artifact
      Dapp::Artifact.new(project: nil, config: openstruct_config)
    end
  end
end
