module SpecHelpers
  module Application
    def application_build!
      application.build_and_fixate!
    end

    def application
      @application || application_renew
    end

    def application_renew
      @application = begin
        options = { conf: config, opts: opts }
        Dapp::Application.new(**options)
      end
    end

    def application_rebuild!
      application_renew
      application_build!
    end

    def config
      raise
    end

    def opts
      { log_quiet: true, build_dir: '' }
    end

    def stages_names
      @stages ||= stages.keys.reverse
    end

    def stages
      hash = {}
      s = application.last_stage
      while s.respond_to? :prev_stage
        hash[s.send(:name)] = s
        s = s.prev_stage
      end
      hash
    end

    def stage(stage_name)
      stages[stage_name]
    end

    def stages_signatures
      stages.values.map { |s| [:"#{s.send(:name)}", s.send(:signature)] }.to_h
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

    def stub_docker_image
      images_cash = []
      stub_instance(Dapp::DockerImage) do |instance|
        allow(instance).to receive(:build!)
        allow(instance).to receive(:exist?)  { images_cash.include? instance.name }
        allow(instance).to receive(:tag!)    { images_cash << instance.name }
        allow(instance).to receive(:pull!)   { images_cash << instance.name }
        allow(instance).to receive(:rmi!)    { images_cash.delete(instance.name) }
        allow(instance).to receive(:id)
      end
    end

    def stub_application
      method_new = Dapp::Application.method(:new)

      application = class_double(Dapp::Application).as_stubbed_const
      allow(application).to receive(:new) do |*args, &block|
        args.first[:conf] = args.first[:conf].to_h.empty? ? RecursiveOpenStruct.new(home_path: '') : args.first[:conf] if args.first.is_a? Hash

        method_new.call(*args, &block).tap do |instance|
          allow(instance).to receive(:build_path) { |*args| Pathname(File.absolute_path(File.join(*args))) }
          allow(instance).to receive(:container_build_path) { |*args| instance.build_path(*args) }
          allow(instance).to receive(:home_path)  { |*args| Pathname(File.absolute_path(File.join(*args))) }
          allow(instance).to receive(:filelock)
        end
      end
    end
  end
end
