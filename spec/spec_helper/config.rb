module SpecHelper
  # Config
  module Config
    def dappfile(&blk)
      @dappfile = ConfigDsl.new.instance_eval(&blk).config
    end

    def dimg_by_name(name)
      dimgs_by_name[name] || raise
    end

    def dimgs_by_name
      dimgs.map { |dimg| [dimg._name, dimg] }.to_h
    end

    def dimg
      dimgs.first
    end

    def dimgs
      Dapp::Config::DimgGroupMain.new(dapp: stubbed_dapp).tap do |config|
        config.instance_eval(@dappfile) unless @dappfile.nil?
      end._dimg
    end

    def stubbed_dapp
      instance_double(Dapp::Dapp).tap do |instance|
        allow(instance).to receive(:name) { File.basename(Dir.getwd) }
        allow(instance).to receive(:log_warning)
      end
    end

    def expect_array_attribute(attribute, dappfile_block)
      dappfile_block.call do
        send(attribute, 'value')
      end

      yield('value')

      dappfile_block.call do
        send(attribute, 'value4')
        send(attribute, 'value1', 'value2', 'value3')
      end

      yield('value4', 'value1', 'value2', 'value3')
    end

    class ConfigDsl
      def initialize
        @config = []
      end

      def config
        @config.join
      end

      def method_missing(name, *args, &blk)
        line("#{name}(#{args.map(&:inspect).join(', ')}) #{'do' if block_given?}")
        if block_given?
          with_indent(&blk)
          line('end')
        end
        self
      end

      def with_indent
        next_indent
        yield if block_given?
        prev_indent
      end

      def line(msg)
        @config << "#{'  ' * (@indent ||= 0)}#{msg}\n"
      end

      def next_indent
        @indent += 1
      end

      def prev_indent
        @indent -= 1
      end
    end
  end # Config
end # SpecHelper
