module Dapp
  module Dimg
    module DockerRegistry
      class Dimg < Base
        def dimgstages_tags
          tags.select { |tag| tag.start_with?('dimgstage') }
        end

        def dimg_tags(dimg_name)
          with_repo_suffix(dimg_name.to_s) { tags }
        rescue Exception::Registry => e
          raise unless e.net_status[:code] == :no_such_dimg
          []
        end

        def nameless_dimg_tags
          tags.select { |tag| !tag.start_with?('dimgstage') }
        end

        def tags
          (@tags ||= {})[repo_suffix] ||= super
        end

        def image_id(tag, extra_repo_suffix = nil)
          with_repo_suffix(extra_repo_suffix.to_s) { super(tag) }
        end

        def image_parent_id(tag, extra_repo_suffix = nil)
          with_repo_suffix(extra_repo_suffix.to_s) { super(tag) }
        end

        def image_labels(tag, extra_repo_suffix = nil)
          with_repo_suffix(extra_repo_suffix.to_s) { super(tag) }
        end

        def image_delete(tag, extra_repo_suffix = nil)
          with_repo_suffix(extra_repo_suffix.to_s) { super(tag) }
        end

        def image_history(tag, extra_repo_suffix = nil)
          with_repo_suffix(extra_repo_suffix.to_s) do
            (@image_history ||= {})[[repo_suffix, tag]] ||= super(tag)
          end
        end

        protected

        def with_repo_suffix(extra_repo_suffix)
          old_repo_suffix = repo_suffix
          @repo_suffix = File.join(repo_suffix, extra_repo_suffix)
          yield
        ensure
          @repo_suffix = old_repo_suffix
        end
      end
    end # DockerRegistry
  end # Dimg
end # Dapp
