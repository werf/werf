module Dapp
  module Deployment
    module Mod
      module Jobs
        [:bootstrap, :before_apply_job].each do |job|
          define_method :"to_kube_#{job}_pods" do |repo, image_version|
            to_kube_default_job_pods(job, repo, image_version)
          end
        end

        def to_kube_default_job_pods(directive, repo, image_version)
          return {} if (directive_config = config.public_send("_#{directive}")).empty?
          {}.tap do |hash|
            hash[name(directive)] = {}.tap do |pod|
              pod['metadata'] = {}.tap do |metadata|
                metadata['name']   = name(directive)
                metadata['labels'] = kube.labels
              end
              pod['spec'] = {}.tap do |spec|
                spec['restartPolicy'] = 'Never'
                spec['containers'] = [].tap do |containers|
                  containers << {}.tap do |container|
                    envs = [environment, secret_environment]
                             .select { |env| !env.empty? }
                             .map { |h| h.map { |k, v| { name: k, value: v } } }
                             .flatten

                    container['imagePullPolicy'] = 'Always'
                    container['image']           = [repo, [directive_config._dimg || config._dimg, image_version].compact.join('-')].join(':')
                    container['name']            = name(directive)
                    container['command']         = directive_config._run unless directive_config._run.empty?
                    container['env']             = envs unless envs.empty?
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
