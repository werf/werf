module SpecHelper
  # Config
  module Config
    def dappfile(&blk)
      @dappfile = ConfigDsl.new.instance_eval(&blk).config
    end

    def dimgs_by_name
      dimgs.map { |dimg| [dimg._name, dimg] }.to_h
    end

    def dimg
      dimgs.first
    end

    def dimgs
      Dapp::Config::DimgGroupMain.new(dappfile_path: File.join(Dir.getwd, 'Dappfile'), project: stubbed_project).tap do |config|
        config.instance_eval(@dappfile) unless @dappfile.nil?
      end._dimgs
    end

    def stubbed_project
      instance_double(Dapp::Project).tap do |instance|
        allow(instance).to receive(:name) { File.basename(Dir.getwd) }
        allow(instance).to receive(:log_warning)
      end
    end

    class ConfigDsl
      def initialize
        @config = []
      end

      def config
        @config.join
      end

      def method_missing(name, *args, &blk)
        @config << line("#{name}(#{args.map(&:inspect).join(', ')}) #{ 'do' if block_given? }")
        if block_given?
          with_indent(&blk)
          @config << line("end")
        end
        self
      end

      def with_indent
        next_indent
        yield if block_given?
        prev_indent
      end

      def line(msg)
        "#{'  ' * (@indent ||= 0)}#{msg}\n"
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