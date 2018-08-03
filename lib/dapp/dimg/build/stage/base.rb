module Dapp
  module Dimg
    module Build
      module Stage
        class Base
          include Helper::Sha256
          include Helper::Trivia
          include Mod::Logging

          attr_accessor :prev_stage, :next_stage
          attr_reader :dimg

          def initialize(dimg, next_stage)
            @dimg = dimg

            @next_stage = next_stage
            @next_stage.prev_stage = self
          end

          def get_ruby2go_state_hash
            {}.tap {|hash|
              hash["Image"] = {
                "Labels" => image.labels,
              }

              if prev_stage
                hash["PrevStage"] = prev_stage.get_ruby2go_state_hash
              end
            }
          end

          def set_ruby2go_state_hash(hash)
          end

          # rubocop:disable Metrics/PerceivedComplexity
          def build_lock!
            return yield if dimg.dapp.dry_run?

            try_lock = proc do
              next yield unless image_should_be_locked?

              no_lock = false

              dimg.dapp.lock("#{dimg.dapp.name}.image.#{image.name}") do
                image.reset_image_inspect

                if image_should_be_locked?
                  yield
                else
                  no_lock = true
                end
              end

              yield if no_lock
            end

            if prev_stage
              prev_stage.build_lock! { try_lock.call }
            else
              try_lock.call
            end
          end
          # rubocop:enable Metrics/MethodLength, Metrics/PerceivedComplexity

          def build!
            prev_stage.build! if prev_stage
            renew             if should_be_renewed?
            image_build
          end

          def save_in_cache!
            prev_stage.save_in_cache! if prev_stage
            return unless should_be_tagged?
            image.save_in_cache! unless dimg.dapp.dry_run?
          end

          def image
            @image ||= begin
              if empty?
                prev_stage.image
              else
                Image::Stage.image_by_name(name: image_name, from: from_image, dapp: dimg.dapp)
              end
            end
          end

          def prepare_image
            return if dimg.dapp.dry_run?

            image_add_mounts

            image.add_service_change_label dapp: dimg.stage_dapp_label
            image.add_service_change_label 'dapp-version'.to_sym => ::Dapp::VERSION
            image.add_service_change_label 'dapp-cache-version'.to_sym => ::Dapp::BUILD_CACHE_VERSION
            image.add_service_change_label 'dapp-dev-mode'.to_sym => true if dimg.dev_mode?

            if dimg.dapp.ssh_auth_sock
              image.add_volume "#{dimg.dapp.ssh_auth_sock}:/tmp/dapp-ssh-agent"
              image.add_env SSH_AUTH_SOCK: '/tmp/dapp-ssh-agent'
            end

            yield if block_given?
          end

          def image_add_mounts
            [:tmp_dir, :build_dir].each do |type|
              next if (mounts = adding_mounts_by_type(type)).empty?

              mounts.each do |path|
                absolute_path = File.expand_path(File.join('/', path))
                tmp_path = dimg.send(type, 'mount', absolute_path[1..-1]).tap(&:mkpath)
                image.add_volume "#{tmp_path}:#{absolute_path}"
              end

              image.add_service_change_label :"dapp-mount-#{type.to_s.tr('_', '-')}" => mounts.join(';')
            end

            image_add_custom_mounts
          end

          def image_add_custom_mounts
            adding_custom_dir_mounts.each do |from, to_pathes|
              FileUtils.mkdir_p(from)
              to_pathes.tap(&:uniq!).map { |to_path| image.add_volume "#{from}:#{to_path}" }
              image.add_service_change_label :"dapp-mount-custom-dir-#{from.gsub('/', '--')}" => to_pathes.join(';')
            end
          end

          def adding_mounts_by_type(type)
            (config_mounts_by_type(type) + labels_mounts_by_type(type)).uniq
          end

          def adding_custom_dir_mounts
            config_custom_dir_mounts.in_depth_merge(labels_custom_dir_mounts)
          end

          def config_custom_dir_mounts
            dimg.config._custom_dir_mount.reduce({}) do |mounts, mount|
              from_path = File.expand_path(mount._from)
              mounts[from_path] ||= []
              mounts[from_path] << mount._to
              mounts
            end
          end

          def config_mounts_by_type(type)
            dimg.config.public_send("_#{type}_mount").map(&:_to)
          end

          def labels_mounts_by_type(type)
            from_image.labels.select { |l, _| l == "dapp-mount-#{type.to_s.tr('_', '-')}" }.map { |_, value| value.split(';') }.flatten
          end

          def labels_custom_dir_mounts
            from_image.labels.map do |label, value|
              next unless label =~ /dapp-mount-custom-dir-(?<from>.+)/
              [File.expand_path(Regexp.last_match(:from).gsub('--', '/')), value.split(';')]
            end.compact.to_h
          end

          def empty?
            dependencies_empty?
          end

          def dependencies_empty?
            dependencies.flatten.compact.delete_if { |val| val.respond_to?(:empty?) && val.empty? }.empty?
          end

          def signature
            if empty?
              prev_stage.signature
            else
              args = []
              args << prev_stage.signature unless prev_stage.nil?
              args << dimg.build_cache_version
              args << builder_checksum
              args.concat(dependencies.flatten)

              hashsum args
            end
          end

          def name
            class_to_lowercase.to_sym
          end

          def builder_checksum
          end

          def git_artifacts_dependencies
            dimg.git_artifacts.map { |git_artifact| git_artifact.stage_dependencies_checksum(self) }
          end

          def layer_commit(git_artifact)
            commits[git_artifact] ||= begin
              if image.built?
                image.labels[dimg.dapp.dimgstage_g_a_commit_label(git_artifact.paramshash)]
              elsif g_a_stage? && !empty?
                git_artifact.latest_commit
              elsif prev_stage
                prev_stage.layer_commit(git_artifact)
              end
            end
          end

          def dependencies
            []
          end

          def dependencies_discard
            @dependencies = nil
          end

          def artifact?
            false
          end

          def g_a_stage?
            false
          end

          protected

          def renew
            commits_discard
            image_reset
            image_untag! if image_should_be_untagged?
          end

          def image_reset
            Image::Stage.image_reset(image_name)
            @image = nil
          end

          def image_untag!
            return if dimg.dapp.dry_run?
            image.untag!
          end

          def image_build
            prepare_image if image_should_be_prepared?

            introspect_image_before_build if image_should_be_introspected_before_build?

            log_image_build do
              dimg.dapp.log_process(log_name,
                                    process: dimg.dapp.t(code: 'status.process.building'),
                                    short: should_not_be_detailed?) { image.build! }
            end unless empty?

            introspect_image_after_build if image_should_be_introspected_after_build?
          end

          def introspect_image_before_build
            introspect_image_default(from_image)
          end

          def introspect_image_after_build
            introspect_image_default(image)
          end

          def introspect_image_default(introspected_image)
            if introspected_image.built?
              introspected_image.introspect!
            else
              dimg.dapp.log_warning(desc: { code: :introspect_image_impossible, data: { name: name } })
            end
          end

          def should_be_tagged?
            image.built? && !image.tagged?
          end

          def image_should_be_locked?
            !(empty? || image.tagged? || should_be_not_present?)
          end

          def image_should_be_prepared?
            (!image.built? && !should_be_not_present? || image_should_be_introspected? && image.tagged?) && !dimg.dapp.dry_run?
          end

          def should_be_renewed?
            current_or_related_image_should_be_untagged?
          end

          def image_should_be_untagged?
            image.tagged? && current_or_related_image_should_be_untagged?
          end

          def current_or_related_image_should_be_untagged?
            @current_or_related_image_should_be_untagged ||= begin
              image_should_be_untagged_condition || begin
                if prev_stage.nil?
                  false
                else
                  prev_stage.current_or_related_image_should_be_untagged?
                end
              end
            end
          end

          def image_should_be_untagged_condition
            return false unless image.tagged?
            dimg.git_artifacts.any? do |git_artifact|
              !git_artifact.repo.commit_exists? layer_commit(git_artifact)
            end
          end

          def should_be_not_present?
            return false if next_stage.nil?
            !current_or_related_image_should_be_untagged? && (next_stage.image.tagged? || next_stage.should_be_not_present?)
          end

          def image_name
            format(dimg.stage_cache_format, signature: signature)
          end

          def from_image
            prev_stage.image if prev_stage || begin
              raise Error::Build, code: :from_image_required
            end
          end

          def change_options
            @change_options ||= begin
              dimg.config._docker._change_options.to_h.delete_if do |_, val|
                val.nil? || (val.respond_to?(:empty?) && val.empty?)
              end
            end
          end

          def commits_discard
            @commits = nil
          end

          def commits
            @commits ||= {}
          end
        end # Base
      end # Stage
    end # Build
  end # Dimg
end # Dapp
