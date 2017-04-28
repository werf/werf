module Dapp
  module Deployment
    class App
      include Namespace
      include SystemEnvironments

      attr_reader :app_config
      attr_reader :deployment

      def initialize(app_config:, deployment:)
        @app_config = app_config
        @deployment = deployment
      end

      def dimg
        deployment.dimgs.find { |dimg| dimg.config._name == app_config._dimg }
      end

      def kube
        @kube ||= KubeApp.new(self)
      end

      [:name, :expose, :bootstrap, :migrate, :run].each do |directive|
        define_method directive do
          app_config.public_send("_#{directive}")
        end
      end

      def to_kube_deployments
        # NOTICE: Не нужно укаывать ApiVersion и Kind
        {
          "hello-backend" => {
            "metadata"=>{
              "name"=>"hello-backend",
              "labels"=>{
                "service"=>"hello-backend"
              }
            },
            "spec"=>{
              "replicas"=>1,
              "template"=>{
                "metadata"=>{
                  "labels"=>{
                    "service"=>"hello-backend"
                  }
                },
                "spec"=>{
                  "containers"=>[
                    {
                      "command"=>['bash'],
                      'args' => ['-lec', "while true ; do date ; sleep 1 ; done"],
                      "env"=>[{"name"=>"TEST_VAR1", "value"=>"value1"}, {"name"=>"TEST_VAR2", "value"=>"value2"}],
                      "image"=>"ubuntu:16.04",
                      "imagePullPolicy"=>"Always",
                      "name"=>"hello-web",
                      "ports"=>[{"containerPort"=>8080, "name"=>"test-web", "protocol"=>"TCP"}]
                    }
                  ]
                }
              }
            }
          }
        }
      end

      def to_kube_services
        # NOTICE: Не нужно указывать ApiVersion и Kind
        {}
      end
    end
  end
end
