module Dapp
  module Build
    class Chef < Base
      [:infra_install, :infra_setup, :app_install, :app_setup].each do |m|
        define_method(:"#{m}_do") do |image|
          prepare_recipes(m) unless chef_cash_file?(m)
          image.build_cmd! "chef-apply #{container_build_path("#{m}_recipes")}"
          image.build_opts! volume: '/opt/chefdk:/opt/chefdk'
        end

        define_method(:"#{m}_signature_do") { hashsum chef_cash_file_sum(m) }
      end

      private

      def chef_path(*path)
        path.compact.inject(build_path('chef'), &:+).expand_path.tap do |p|
          FileUtils.mkdir_p p.parent
        end
      end

      def container_chef_path(*path)
        path.compact.inject(container_build_path('chef'), &:+).expand_path
      end

      def chef_cash_file_sum(stage)
        if berksfile_lock?
          prepare_recipes(stage) unless chef_cash_file?(stage)
          hashsum(File.read(chef_cash_file_path(stage)))
        end
      end

      def chef_cash_file?(stage)
        File.exist?(chef_cash_file_path(stage))
      end

      def chef_cash_file_path(stage)
        home_path("#{stage}.#{berksfile_lock_sum}")
      end

      def berksfile_lock_sum
        hashsum(File.read(berksfile_lock_path)) if berksfile_lock?
      end

      def berksfile_lock?
        File.exist?(berksfile_lock_path)
      end

      def berksfile_lock_path
        home_path('Berksfile.lock')
      end

      def prepare_recipes(stage)
        # vendor
        shellout("berks vendor #{chef_path('vendor')}")

        # stage recipes
        recipes = []
        reg = /#{stage}*.rb/
        conf[:mdapps].each { |mdapp| Dir[chef_path(mdapp, 'recipes', reg)].each { |path| recipes << path } }
        recipes += Dir[base_dir('recipes', reg)]
        recipes.map! { |path| Pathname(path) }

        # modified chef_cash_file_path
        File.open(chef_cash_file_path(stage), 'w') {|f| f.puts(hashsum(recipes.map(&:read))) }

        # number and cp recipes
        FileUtils.rm_rf(build_path("#{stage}_recipes"))
        recipes.each_with_index { |file, index| FileUtils.copy(file.path, build_path("#{stage}_recipes", "#{index}_#{file.name}")) }
      end
    end
  end
end

