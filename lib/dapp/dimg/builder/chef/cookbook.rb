module Dapp
  module Dimg
    # Локальный cookbook сборщик с правилами
    # сборки dimg образов для chef сборщика.
    class Builder::Chef::Cookbook
      CACHE_VERSION = 1

      CHECKSUM_PATTERNS = %w(
        attributes/**/*
        recipes/**/*
        files/**/*
        templates/**/*
      ).freeze

      attr_reader :builder

      attr_reader :path
      attr_reader :vendor_path
      attr_reader :berksfile
      attr_reader :metadata
      attr_reader :enabled_recipes
      attr_reader :enabled_modules

      # rubocop:disable Metrics/ParameterLists
      def initialize(builder, path:, berksfile:, metadata:, enabled_recipes:, enabled_modules:)
        @builder = builder
        @path = Pathname.new(path)
        @berksfile = berksfile
        @metadata = metadata
        @enabled_recipes = enabled_recipes
        @enabled_modules = enabled_modules
      end
      # rubocop:enable Metrics/ParameterLists

      def name
        metadata.name
      end

      def local_paths
        @local_paths ||= [].tap do |paths|
          paths << path
          paths.push(*berksfile.local_cookbooks.values.map {|cookbook| cookbook[:path]})
          paths.uniq!
        end
      end

      # TODO: Сookbook'и зависимости на локальной файловой системе
      # TODO: должны обрабатываться рекурсивно. Чтобы контрольная сумма
      # TODO: учитывала все зависимости всех зависимых локальных cookbook'ов.
      def checksum
        @checksum ||= begin
          all_local_paths = local_paths
                            .product(CHECKSUM_PATTERNS)
                            .map {|cb, dir| Dir[cb.join(dir)]}
                            .flatten
                            .map(&Pathname.method(:new))

          builder.dimg.hashsum [
            CACHE_VERSION,
            builder.dimg.paths_content_hashsum(all_local_paths),
            *all_local_paths.map {|p| p.relative_path_from(path).to_s}.sort,
            berksfile.dump,
            metadata.dump
          ].compact
        end
      end

      def vendor_path
        builder.dimg.dapp.lock(_vendor_lock_name, default_timeout: 120) do
          vendor! unless _vendor_path.join('.created_at').exist?
        end
        _vendor_path
      end

      # rubocop:disable Metrics/AbcSize
      def vendor!
        volumes_from = [builder.dimg.dapp.base_container, builder.chefdk_container]

        builder.dimg.dapp.log_secondary_process(builder.dimg.dapp.t(code: _vendor_process_name)) do
          volumes_from = [builder.dimg.dapp.base_container, builder.chefdk_container]

          tmp_berksfile_path = builder.dimg.tmp_path.join("Berksfile.#{SecureRandom.uuid}")
          File.open(tmp_berksfile_path, 'w') {|file| file.write berksfile.dump}

          tmp_metadata_path = builder.dimg.tmp_path.join("metadata.rb.#{SecureRandom.uuid}")
          File.open(tmp_metadata_path, 'w') {|file| file.write metadata.dump}

          vendor_commands = [
            "#{builder.dimg.dapp.mkdir_bin} -p ~/.ssh",
            'echo "Host *" >> ~/.ssh/config',
            'echo "    StrictHostKeyChecking no" >> ~/.ssh/config',
            *local_paths.map {|path| "#{builder.dimg.dapp.rsync_bin} --archive --relative #{path} /tmp/local_cookbooks"},
            "cd /tmp/local_cookbooks/#{path}",
            "cp #{tmp_berksfile_path} Berksfile",
            "cp #{tmp_metadata_path} metadata.rb",
            '/.dapp/deps/chefdk/bin/berks vendor /tmp/cookbooks',
            ["#{builder.dimg.dapp.find_bin} /tmp/cookbooks -type d -exec #{builder.dimg.dapp.bash_bin} -ec '",
             "#{builder.dimg.dapp.install_bin} -o #{Process.uid} -g #{Process.gid} ",
             "--mode $(#{builder.dimg.dapp.stat_bin} -c %a {}) -d ",
             "#{_vendor_path}/$(echo {} | #{builder.dimg.dapp.sed_bin} -e \"s/^\\/tmp\\/cookbooks//\")' \\;"].join,
            ["#{builder.dimg.dapp.find_bin} /tmp/cookbooks -type f -exec #{builder.dimg.dapp.bash_bin} -ec '",
             "#{builder.dimg.dapp.install_bin} -o #{Process.uid} -g #{Process.gid} ",
             "--mode $(#{builder.dimg.dapp.stat_bin} -c %a {}) {} ",
             "#{_vendor_path}/$(echo {} | #{builder.dimg.dapp.sed_bin} -e \"s/\\/tmp\\/cookbooks//\")' \\;"].join,
            ["#{builder.dimg.dapp.install_bin} -o #{Process.uid} -g #{Process.gid} ",
             "--mode 0644 <(#{builder.dimg.dapp.date_bin} +%s.%N) #{_vendor_path.join('.created_at')}"].join
          ]

          builder.dimg.dapp.shellout!(
            [ 'docker run --rm',
              volumes_from.map {|container| "--volumes-from #{container}"}.join(' '),
              *local_paths.map {|path| "--volume #{path}:#{path}"},
              "--volume #{builder.dimg.tmp_path}:#{builder.dimg.tmp_path}",
              ("--volume #{builder.dimg.dapp.ssh_auth_sock}:/tmp/dapp-ssh-agent" if builder.dimg.dapp.ssh_auth_sock),
              "--volume #{_vendor_path.tap(&:mkpath)}:#{_vendor_path}",
              ('--env SSH_AUTH_SOCK=/tmp/dapp-ssh-agent' if builder.dimg.dapp.ssh_auth_sock),
              ''.tap do |cmd|
                cmd << "dappdeps/berksdeps:0.1.0 #{builder.dimg.dapp.bash_bin}"
                cmd << " -ec '#{builder.dimg.dapp.shellout_pack(vendor_commands.join(' && '))}'"
              end ].compact.join(' '),
            log_verbose: builder.dimg.dapp.log_verbose?
          )
        end
      end
      # rubocop:enable Metrics/AbcSize

      def stage_cookbooks_path(stage)
        _stage_cookbooks_path(stage).tap do |_cookbooks_path|
          @stage_cookbooks_installed ||= {}
          @stage_cookbooks_installed[stage] ||= true.tap {install_stage_cookbooks(stage)}
        end
      end

      def stage_enabled_modules(stage)
        @stage_enabled_modules ||= {}
        @stage_enabled_modules[stage] ||= enabled_modules.select {|cookbook| stage_entry_exist?(stage, cookbook, stage)}
      end

      def stage_enabled_recipes(stage)
        @stage_enabled_recipes ||= {}
        @stage_enabled_recipes[stage] ||= enabled_recipes.select {|recipe| stage_entry_exist?(stage, name, recipe)}
      end

      def stage_checksum(stage)
        paths = Dir[stage_cookbooks_path(stage).join('**/*')].map(&Pathname.method(:new))

        builder.dimg.hashsum [
          builder.dimg.paths_content_hashsum(paths),
          *paths.map {|p| p.relative_path_from(stage_cookbooks_path(stage)).to_s}.sort,
          stage == :before_install ? builder.chefdk_image : nil
        ].compact
      end

      protected

      def _vendor_path
        builder.dimg.build_path.join('cookbooks', checksum)
      end

      def _vendor_process_name
        'process.vendoring_builder_cookbooks'
      end

      def _vendor_lock_name
        "#{builder.dimg.dapp.name}.cookbooks.#{checksum}"
      end

      def stage_entry_exist?(stage, cookbook, entrypoint)
        stage_cookbooks_path(stage).join(cookbook, 'recipes', "#{entrypoint}.rb").exist?
      end

      def install_stage_cookbooks(stage)
        _stage_cookbooks_path(stage).mkpath

        stage_install_paths(stage).each do |cookbook_path, paths|
          cookbook = cookbook_path.basename.to_s

          paths.each do |from, to|
            if from.nil?
              to_path = _stage_cookbooks_path(stage).join(cookbook, 'recipes/void.rb')
              to_path.parent.mkpath
              FileUtils.touch to_path
            else
              from_path = cookbook_path.join(from)
              to_path = _stage_cookbooks_path(stage).join(cookbook, to)
              if from_path.directory? && to_path.exist?
                Dir[from_path.join('**/*')]
                  .map(&Pathname.method(:new))
                  .each do |from_subpath|
                    to_subpath = to_path.join(from_subpath.relative_path_from(from_path))
                    if to_subpath.exist?
                      raise(Error::Chef,
                            code: :stage_path_overlap,
                            data: { stage: stage,
                                    cookbook: cookbook,
                                    from: from_subpath.relative_path_from(cookbook_path),
                                    to: to_subpath.relative_path_from(_stage_cookbooks_path(stage).join(cookbook)) })
                    end

                    to_subpath.parent.mkpath
                    FileUtils.cp_r from_subpath, to_subpath
                  end
              else
                to_path.parent.mkpath
                FileUtils.cp_r from_path, to_path
              end
            end
          end
        end
      end

      # rubocop:disable Metrics/PerceivedComplexity, Metrics/CyclomaticComplexity, Metrics/MethodLength
      def stage_install_paths(stage)
        select_existing_paths = proc do |cookbook_path, paths|
          paths.select {|from, _| cookbook_path.join(from).exist?}
        end

        common_paths = [['metadata.json', 'metadata.json']]

        Dir[vendor_path.join('*')]
          .map(&Pathname.method(:new))
          .map do |cookbook_path|
          cookbook_name = File.basename cookbook_path
          is_builder = (cookbook_name == name)
          is_dimod = cookbook_name.start_with? 'dimod-'
          dimod_enabled = is_dimod && enabled_modules.include?(cookbook_name)

          paths = if is_builder
            common_dapp_paths = select_existing_paths.call(
              cookbook_path,
              [
                *common_paths,
                ["files/#{stage}/common", 'files/default'],
                ["templates/#{stage}/common", 'templates/default'],
                *enabled_recipes.flat_map do |recipe|
                  [["files/#{stage}/#{recipe}", 'files/default'],
                   ["templates/#{stage}/#{recipe}", 'templates/default']]
                end
              ]
            )

            recipe_paths = enabled_recipes
                           .map {|recipe| ["recipes/#{stage}/#{recipe}.rb", "recipes/#{recipe}.rb"]}
                           .select {|from, _| cookbook_path.join(from).exist?}

            if recipe_paths.any?
              [*recipe_paths, *common_dapp_paths]
            else
              [nil, *common_dapp_paths]
            end
          elsif is_dimod && dimod_enabled
            common_dimod_paths = select_existing_paths.call(
              cookbook_path,
              [
                *common_paths,
                ["files/#{stage}", 'files/default'],
                ['files/common', 'files/default'],
                ["templates/#{stage}", 'templates/default'],
                ['templates/common', 'templates/default'],
                ["attributes/#{stage}.rb", "attributes/#{stage}.rb"],
                ['attributes/common.rb', 'attributes/common.rb']
              ]
            )

            recipe_path = "recipes/#{stage}.rb"
            if cookbook_path.join(recipe_path).exist?
              [[recipe_path, recipe_path], *common_dimod_paths]
            else
              [nil, *common_dimod_paths]
            end
          elsif !is_dimod
            [['.', '.']]
          end

          [cookbook_path, paths] if paths && paths.any?
        end.compact
      end
      # rubocop:enable Metrics/PerceivedComplexity, Metrics/CyclomaticComplexity, Metrics/MethodLength

      def _stage_cookbooks_path(stage)
        builder.stage_build_path(stage).join('cookbooks')
      end
    end # Builder::Chef::Cookbook
  end # Dimg
end # Dapp
