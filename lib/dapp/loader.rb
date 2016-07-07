module Dapp
  class Loader
    class << self
      def process_directory(path, pattern = '*', **options)
        dappfiles_paths(path, pattern).map { |dappfile_path| process_file(dappfile_path, app_filter: pattern, **options) }.flatten
      end

      def process_file(dappfile_path, app_filter: '*', **options)
        Config::Main.new(dappfile_path: dappfile_path, **options) do |conf|
          conf.log "Processing dappfile '#{dappfile_path}'"
          conf.instance_eval File.read(dappfile_path), dappfile_path
        end.apps # TODO app_fitter
      end

      def dappfiles_paths(path, pattern = '*')
        pattern.split('-').instance_eval { count.downto(1).map { |n| slice(0, n).join('-') } }
            .map { |p| Dir.glob(File.join([path, p, 'Dappfile'].compact)) }.find(&:any?) || []
      end
    end
  end
end
