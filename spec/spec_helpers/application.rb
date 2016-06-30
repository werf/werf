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
        options = { conf: config.dup, opts: opts }
        Dapp::Application.new(**options)
      end
    end

    def config
      raise
    end

    def opts
      raise
    end

    def stages_names
      @stages ||= stages.keys.reverse
    end

    def stages
      stgs = {}
      s = application.last_stage
      while s.respond_to? :prev_stage
        stgs[s.send(:name)] = s
        s = s.prev_stage
      end
      stgs
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
      method_new = Dapp::DockerImage.method(:new)

      docker_image = class_double(Dapp::DockerImage).as_stubbed_const
      allow(docker_image).to receive(:new) do |*args, &block|
        method_new.call(*args, &block).tap do |instance|
          allow(instance).to receive(:build!)
          allow(instance).to receive(:exist?)  { images_cash.include? instance.name }
          allow(instance).to receive(:tag!)    { images_cash << instance.name }
        end
      end
    end

    def stub_application
      @application = instance_double('Dapp::Application')
      allow(@application).to receive(:build_path) do |*args|
        File.absolute_path(File.join(*args))
      end
      allow(@application).to receive(:container_build_path) do |*args|
        File.absolute_path(File.join(*args))
      end
      allow(@application).to receive(:home_path).and_return('')
      allow(@application).to receive(:shellout!) do |*args, **kwargs|
        shellout(*args, **kwargs)
      end
      allow(@application).to receive(:filelock).and_yield
    end
  end
end
