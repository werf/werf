module SpecHelper
  module Application
    def application_build!
      application.build!
    end

    def application
      @application || application_renew
    end

    def application_renew
      @openstruct_config = nil
      @application = begin
        options = { config: openstruct_config, cli_options: cli_options }
        Dapp::Application.new(**options)
      end
    end

    def application_rebuild!
      application_renew
      application_build!
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
                                _artifact: [],
                                _chef: { _modules: [] },
                                _shell: { _infra_install: [], _infra_setup: [], _install: [], _setup: [] },
                                _docker: { _from: :'ubuntu:14.04', _change_options: {} },
                                _git_artifact: { _local: [], _remote: [] },
                                _install_dependencies: [],
                                _setup_dependencies: []))
    end

    def cli_options
      default_cli_options
    end

    def default_cli_options
      { log_quiet: true, log_indent: 0 }
    end

    def stages
      hash = {}
      s = application.send(:last_stage)
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

    # rubocop:disable Metrics/AbcSize
    def stub_application
      method_new = Dapp::Application.method(:new)

      application = class_double(Dapp::Application).as_stubbed_const
      allow(application).to receive(:new) do |*args, &block|
        method_new.call(*args, &block).tap do |instance|
          allow(instance).to receive(:home_path) { |*m_args| Pathname(File.absolute_path(File.join(*m_args))) }
          allow(instance).to receive(:filelock)
        end
      end
    end
    # rubocop:enable Metrics/AbcSize
  end
end
