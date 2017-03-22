module Dapp
  module Dimg
    class Builder::Chef < Builder::Base
      DEFAULT_CHEFDK_IMAGE = 'dappdeps/chefdk:0.17.3-1'.freeze # TODO: config, DSL, DEFAULT_CHEFDK_IMAGE

      %i(before_install install before_setup setup build_artifact).each do |stage|
        define_method("#{stage}?") {!stage_empty?(stage)}

        define_method("#{stage}_checksum") do
          dimg.hashsum [stage_cookbooks_checksum(stage),
                        stage_attributes_raw(stage),
                        *stage_cookbooks_runlist(stage)]
        end

        define_method(stage.to_s) do |image|
          unless stage_empty?(stage)
            image.add_volumes_from(chefdk_container)
            image.add_volume "#{stage_build_path(stage)}:#{container_stage_build_path(stage)}:ro"
            image.add_command ['/.dapp/deps/chefdk/bin/chef-solo',
                               '--legacy-mode',
                               "--config #{container_stage_config_path(stage)}",
                               "--json-attributes #{container_stage_json_attributes_path(stage)}",
                               "--override-runlist #{stage_cookbooks_runlist(stage).join(',')}"].join(' ')
          end
        end
      end

      def before_dimg_should_be_built_check
        super

        %i(before_install install before_setup setup).each do |stage|
          unless stage_empty?(stage) || stage_cookbooks_checksum_path(stage).exist?
            raise Error::Chef, code: :stage_checksum_not_calculated, data: {stage: stage}
          end
        end
      end

      def before_build_check
        %i(before_install install before_setup setup build_artifact).tap do |stages|
          (builder_cookbook.enabled_recipes -
           stages.map {|stage| builder_cookbook.stage_enabled_recipes(stage)}.flatten.uniq).each do |recipe|
            dimg.dapp.log_warning(desc: {code: :recipe_does_not_used, data: {recipe: recipe}})
          end
        end
      end

      def chefdk_image
        DEFAULT_CHEFDK_IMAGE # TODO: config, DSL, DEFAULT_CHEFDK_IMAGE
      end

      def chefdk_container_name # FIXME: hashsum(image) or dockersafe()
        chefdk_image.tr('/', '_').tr(':', '_')
      end

      def chefdk_container
        @chefdk_container ||= begin
          if dimg.dapp.shellout("docker inspect #{chefdk_container_name}").exitstatus.nonzero?
            dimg.dapp.log_secondary_process(dimg.dapp.t(code: 'process.chefdk_container_creating'), short: true) do
              dimg.dapp.shellout!(
                ['docker create',
                 "--name #{chefdk_container_name}",
                 "--volume /.dapp/deps/chefdk #{chefdk_image}"].join(' ')
              )
            end
          end

          chefdk_container_name
        end
      end

      def builder_cookbook
        @builder_cookbook ||= begin
          unless dimg.dapp.builder_cookbook_path.exist?
            raise Error::Chef,
                  code: :builder_cookbook_not_found,
                  data: {path: dimg.dapp.builder_cookbook_path.to_s}
          end

          cookbooks = Marshal.load Marshal.dump(dimg.config._chef._cookbook)

          cookbooks.each do |_name, desc|
            # Получение относительного пути из директории .dapp_chef до указанной зависимости.
            # В Dappfile указываются пути относительно самого Dappfile либо абсолютные пути.
            # В объекте конфига должны лежать абсолютные пути по ключу :path.
            if desc[:path]
              relative_from_cookbook_path = Pathname.new(desc[:path]).relative_path_from(dimg.dapp.builder_cookbook_path).to_s
              desc[:path] = relative_from_cookbook_path
            end
          end

          # Добавление самого cookbook'а builder'а.
          cookbooks[dimg.dapp.name] = {
            name: dimg.dapp.name,
            path: '.'
          }

          berksfile = Berksfile.from_conf(cookbook_path: dimg.dapp.builder_cookbook_path.to_s, cookbooks: cookbooks)
          metadata = CookbookMetadata.from_conf(name: dimg.dapp.name, version: '1.0.0', cookbooks: cookbooks)

          Cookbook.new(self,
                       path: dimg.dapp.builder_cookbook_path,
                       berksfile: berksfile,
                       metadata: metadata,
                       enabled_recipes: dimg.config._chef._recipe,
                       enabled_modules: dimg.config._chef._dimod)
        end
      end

      def stage_attributes(stage)
        dimg.config._chef.send("__#{stage}_attributes")
      end

      def stage_attributes_raw(stage)
        JSON.dump stage_attributes(stage)
      end

      def stage_cookbooks_checksum_path(stage)
        dimg.build_path.join("#{builder_cookbook.checksum}.#{stage}.checksum")
      end

      def stage_cookbooks_checksum(stage)
        if stage_cookbooks_checksum_path(stage).exist?
          stage_cookbooks_checksum_path(stage).read.strip
        else
          checksum = builder_cookbook.stage_checksum(stage)

          stage_cookbooks_checksum_path(stage).tap do |path|
            path.parent.mkpath
            path.write "#{checksum}\n"
          end

          checksum
        end
      end

      # rubocop:disable Metrics/PerceivedComplexity
      def stage_cookbooks_runlist(stage)
        @stage_cookbooks_runlist ||= {}
        @stage_cookbooks_runlist[stage] ||= begin
          res = []

          format_entry = proc do |cookbook, entrypoint|
            entrypoint = 'void' if entrypoint.nil?
            "#{cookbook}::#{entrypoint}"
          end

          builder_cookbook.enabled_modules.map do |cookbook|
            if builder_cookbook.stage_enabled_modules(stage).include? cookbook
              [cookbook, stage]
            else
              [cookbook, nil]
            end
          end.tap {|entries| res.concat entries}

          builder_cookbook.stage_enabled_recipes(stage)
                          .map {|recipe| [builder_cookbook.name, recipe]}
                          .tap do |entries|
            if entries.any?
              res.concat entries
            else
              res << [builder_cookbook.name, nil]
            end
          end

          if res.all? {|_, entrypoint| entrypoint.nil?}
            []
          else
            res.map(&format_entry)
          end
        end
      end
      # rubocop:enable Metrics/PerceivedComplexity

      def stage_empty?(stage)
        stage_cookbooks_runlist(stage).empty?
      end

      def install_chef_solo_stage_config(stage)
        @install_chef_solo_stage_config ||= {}
        @install_chef_solo_stage_config[stage] ||= true.tap do
          stage_build_path(stage).join('config.rb').write [
            "file_cache_path \"/.dapp/chef/cache\"\n",
            "cookbook_path \"#{container_stage_build_path(stage).join('cookbooks')}\"\n"
          ].join
        end
      end

      def container_stage_config_path(stage)
        install_chef_solo_stage_config(stage)
        container_stage_build_path(stage).join('config.rb')
      end

      def install_json_attributes(stage)
        @install_json_attributes ||= {}
        @install_json_attributes[stage] ||= true.tap do
          stage_build_path(stage).join('attributes.json').write "#{stage_attributes_raw(stage)}\n"
        end
      end

      def container_stage_json_attributes_path(stage)
        install_json_attributes(stage)
        container_stage_build_path(stage).join('attributes.json')
      end

      def stage_build_path(stage)
        dimg.tmp_path(dimg.config._name).join(stage.to_s)
      end

      def container_stage_build_path(_stage)
        Pathname.new('/.dapp/chef/build')
      end
    end # Builder::Chef
  end # Dimg
end # Dapp
