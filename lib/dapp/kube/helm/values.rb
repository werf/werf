module Dapp
  module Kube
    class Helm::Values
      TEMPLATE_EMPTY_VALUE = "\\\"-\\\""

      class << self
        def service_values(*a, &b)
          self.new(service_values_hash(*a, &b))
        end

        def service_values_hash(dapp, repo, namespace, docker_tag, fake: false, without_registry: false, disable_warnings: false)
          res = {
            "global" => {
              "namespace" => namespace,
              "dapp" => {
                "name" => dapp.name,
                "repo" => repo,
                "docker_tag" => docker_tag,
              },
              "ci" => ENV.select { |k, _| k.start_with?("CI_") } ,
            }
          }

          ci_info = {
            "is_tag" => false,
            "is_branch" => false,
            "branch" => TEMPLATE_EMPTY_VALUE,
            "tag" => TEMPLATE_EMPTY_VALUE,
            "ref" => TEMPLATE_EMPTY_VALUE
          }
          res["global"]["dapp"]["ci"] = ci_info

          if fake
          elsif ENV["CI_COMMIT_TAG"]
            ci_info["tag"] = ci_info["ref"] = ENV["CI_COMMIT_TAG"]
            ci_info["is_tag"] = true
          elsif ENV["CI_COMMIT_REF_NAME"]
            ci_info["branch"] = ci_info["ref"] = ENV["CI_COMMIT_REF_NAME"]
            ci_info["is_branch"] = true
          elsif dapp.git_path and dapp.git_local_repo.branch != "HEAD"
            ci_info["branch"] = ci_info["ref"] = dapp.git_local_repo.branch
            ci_info["is_branch"] = true
          elsif dapp.git_path
            git = dapp.git_local_repo.send(:git)

            tagref = git.references.find do |r|
              if r.name.start_with?("refs/tags/")
                if r.target.is_a? Rugged::Tag::Annotation
                  r.target.target_id == git.head.target_id
                else
                  r.target_id == git.head.target_id
                end
              end
            end

            if tagref
              tag = tagref.name.partition("refs/tags/").last
            else
              tag = git.head.target_id
            end

            ci_info["tag"] = ci_info["ref"] = tag
            ci_info["is_tag"] = true
          end

          dimgs = dapp.build_configs.map do |config|
            ::Dapp::Dimg::Dimg.new(config: config, dapp: dapp, ignore_git_fetch: true)
          end.uniq do |dimg|
            dimg.config._name
          end

          dimgs.each do |dimg|
            dimg_data = {}
            if dimg.config._name
              res["global"]["dapp"]["is_nameless_dimg"] = false
              res["global"]["dapp"]["dimg"] ||= {}
              res["global"]["dapp"]["dimg"][dimg.config._name] = dimg_data
            else
              res["global"]["dapp"]["is_nameless_dimg"] = true
              res["global"]["dapp"]["dimg"] = dimg_data
            end

            dimg_labels = {}
            docker_image_id = TEMPLATE_EMPTY_VALUE
            unless fake || without_registry
              begin
                dimg_labels = dapp.dimg_registry(repo).image_labels(docker_tag, dimg.config._name)
                docker_image_id = dapp.dimg_registry(repo).image_id(docker_tag, dimg.config._name)
              rescue ::Dapp::Dimg::Error::Registry => err
                unless disable_warnings
                  dapp.log_warning "Registry `#{err.net_status[:data][:registry]}` is not availabble: cannot determine <dimg>.docker_image_id and <dimg>.git.<ga>.commit_id helm values of dimg#{dimg.config._name ? " `#{dimg.config._name}`" : nil}"
                end
              end
            end

            dimg_data["docker_image"] = [[repo, dimg.config._name].compact.join("/"), docker_tag].join(":")
            dimg_data["docker_image_id"] = docker_image_id

            [*dimg.local_git_artifacts, *dimg.remote_git_artifacts].each do |ga|
              if ga.as
                commit_id = dimg_labels[dapp.dimgstage_g_a_commit_label(ga.paramshash)] || TEMPLATE_EMPTY_VALUE

                dimg_data["git"] ||= {}
                dimg_data["git"][ga.as] ||= {}
                dimg_data["git"][ga.as]["commit_id"] = commit_id
              end
            end
          end

          res
        end
      end

      attr_reader :data

      def initialize(data={})
        @data = data
      end

      def as_set_options
        options = {}

        queue = [[nil, data]]

        loop do
          option_key, hash = queue.shift
          break unless hash

          hash.each do |k, v|
            new_option_key = [option_key, k].compact.join(".")
            if v.is_a? Hash
              queue << [new_option_key, v]
            else
              options[new_option_key] = v
            end
          end
        end

        options
      end

      def to_set_options
        as_set_options.map {|k, v| "--set #{k}=#{v}"}
      end
    end # Helm::ServiceValues
  end # Kube
end # Dapp
