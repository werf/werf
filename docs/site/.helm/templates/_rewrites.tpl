# DON'T FORGET TO MAKE CHANGES ALSO IN docs/documentation/.helm/_rewrites.tpl if necessary.
{{- define "rewrites" }}
rewrite ^(/v[\d]+\.[\d]+([^\/]+)?)+/documentation/(.*)$ /documentation$1/$3 permanent;

# 202103
rewrite ^/documentation/[^/]*/?advanced/supported_registry_implementations\.html /documentation/v1.2/advanced/supported_container_registries.html permanent;

#
rewrite ^(/([^/]+/)?)+documentation/advanced/configuration/organizing_configuration\.html /documentation/v1.2/reference/werf_yaml_template_engine.html permanent;
rewrite ^/documentation/advanced/configuration/supported_go_templates\.html$   /documentation/v1.2/reference/werf_yaml_template_engine.html permanent;
rewrite ^/documentation/advanced/configuration/giterminism\.html$   /documentation/v1.2/advanced/giterminism.html permanent;

rewrite ^(/v1\.[01]+(\-[a-z]+)?)+/documentation/quickstart\.html$            /documentation/v1.1/guides/getting_started.html permanent;

rewrite ^/documentation/configuration/introduction\.html$                   /documentation/v1.2/reference/werf_yaml.html permanent;
rewrite ^/documentation/configuration/dockerfile_image\.html$               /documentation/v1.2/reference/werf_yaml.html permanent;
rewrite ^/documentation/configuration/stapel_image/naming\.html$            /documentation/v1.2/reference/werf_yaml.html#image-section permanent;
rewrite ^/documentation/configuration/stapel_image/(.+)\.html$              /documentation/v1.2/advanced/building_images_with_stapel/$1.html permanent;
rewrite ^/documentation/configuration/stapel_image/image_directives\.html$  /documentation/v1.2/reference/werf_yaml.html permanent;
rewrite ^/documentation/configuration/stapel_artifact\.html$                /documentation/v1.2/advanced/building_images_with_stapel/artifact.html permanent;
rewrite ^/documentation/configuration/deploy_into_kubernetes\.html$         /documentation/v1.2/reference/werf_yaml.html#deploy  permanent;
rewrite ^/documentation/configuration/cleanup\.html$                        /documentation/v1.2/reference/werf_yaml.html#cleanup  permanent;

rewrite ^/documentation/reference/build_process\.html$                                        /documentation/v1.2/internals/build_process.html permanent;
rewrite ^/documentation/reference/stages_and_images\.html$                                    /documentation/v1.2/internals/stages_and_storage.html permanent;
rewrite ^/documentation/reference/deploy_process/deploy_into_kubernetes\.html$                /documentation/v1.2/advanced/helm/overview.html permanent;
rewrite ^/documentation/advanced/helm/working_with_secrets\.html$                             /documentation/v1.2/advanced/helm/configuration/secrets.html permanent;

rewrite ^/documentation/reference/deploy_process/working_with_chart_dependencies\.html$       /documentation/v1.2/advanced/helm/configuration/chart_dependencies.html permanent;
rewrite ^/documentation/advanced/helm/working_with_chart_dependencies\.html$                  /documentation/v1.2/advanced/helm/configuration/chart_dependencies.html permanent;

rewrite ^/documentation/reference/cleaning_process\.html$                                     /documentation/v1.2/advanced/cleanup.html permanent;
rewrite ^/documentation/reference/working_with_docker_registries\.html$                       /documentation/v1.2/advanced/supported_registry_implementations.html permanent;
rewrite ^/documentation/reference/ci_cd_workflows_overview\.html$                             /documentation/v1.2/advanced/ci_cd/ci_cd_workflow_basics.html permanent;
rewrite ^/documentation/reference/plugging_into_cicd/overview\.html$                          /documentation/v1.2/internals/how_ci_cd_integration_works/general_overview.html permanent;
rewrite ^/documentation/reference/plugging_into_cicd/gitlab_ci\.html$                         /documentation/v1.2/internals/how_ci_cd_integration_works/gitlab_ci_cd.html permanent;
rewrite ^/documentation/reference/development_and_debug/stage_introspection\.html$            /documentation/v1.2/advanced/development_and_debug/stage_introspection.html permanent;

rewrite ^/documentation/reference/development_and_debug/lint_and_render_chart\.html$          /documentation/v1.2/advanced/development_and_debug/lint_and_render_chart.html permanent;
rewrite ^/documentation/reference/toolbox/slug\.html$                                         /documentation/v1.2/internals/names_slug_algorithm.html permanent;
rewrite ^/documentation/reference/toolbox/ssh\.html$                                          /documentation/v1.2/internals/integration_with_ssh_agent.html permanent;
rewrite ^/documentation/cli/(main|toolbox|other)/([^/]+)\.html$                            /documentation/v1.2/reference/cli/werf_$2.html permanent;
rewrite ^/documentation/cli/management/([^/]+)/([^/]+)\.html$                        /documentation/v1.2/reference/cli/werf_$1_$2.html permanent;
rewrite ^/documentation/cli/management/([^/]+)/([^/]+)/([^/]+)\.html$                /documentation/v1.2/reference/cli/werf_$1_$2_$3.html permanent;
rewrite ^/documentation/cli/management/([^/]+)/([^/]+)/([^/]+)/([^/]+)\.html$        /documentation/v1.2/reference/cli/werf_$1_$2_$3_$4.html permanent;
rewrite ^/documentation/development/stapel\.html$                                    /documentation/v1.2/internals/development/stapel_image.html permanent;
rewrite ^/documentation/guides/installation\.html$                                   /installation.html permanent;
rewrite ^(/v1\.[^01]+(\-[a-z]+)?)+/documentation/guides/(getting_started|deploy_into_kubernetes)+\.html$       /documentation/v1.2//quickstart.html permanent;
rewrite ^/documentation/guides/(getting_started|deploy_into_kubernetes)+\.html$      /documentation/v1.2/quickstart.html permanent;
rewrite ^/documentation/guides/generic_ci_cd_integration\.html$                      /documentation/v1.2/advanced/ci_cd/generic_ci_cd_integration.html permanent;
rewrite ^/documentation/guides/gitlab_ci_cd_integration\.html$                       /documentation/v1.2/advanced/ci_cd/gitlab_ci_cd.html permanent;
rewrite ^/documentation/guides/github_ci_cd_integration\.html$                       /documentation/v1.2/advanced/ci_cd/github_actions.html permanent;
rewrite ^/documentation/guides/advanced_build/(first_application|multi_images|mounts|artifacts)+\.html$    /guides.html permanent;

rewrite ^(/v1.1(\-(alpha|beta|ea|stable)+)?)+/documentation/guides/unsupported_ci_cd_integration\.html$ /documentation/v1.1/guides/generic_ci_cd_integration.html permanent;

rewrite ^/publications_ru\.html$  https://ru.werf.io/publications.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation\.html$  /documentation/ permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/?$  /installation.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/guides/installation\.html$  /installation.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/mounts\.html$  /documentation/v1.1/guides/advanced_build/mounts.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/multi_images\.html$  /documentation/v1.1/guides/advanced_build/multi_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/artifacts\.html$  /documentation/v1.1/guides/advanced_build/artifacts.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/(.+)  /documentation/v1.1/guides/$3 permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/?$  /documentation/v1.1/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/config\.html$  /documentation/v1.1/configuration/introduction.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/artifact\.html$  /documentation/advanced/building_images_with_stapel/artifacts.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/as_layers\.html$  /documentation/v1.1/reference/development_and_debug/as_layers.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/stage_introspection\.html$  /documentation/v1.1/reference/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/(.+)\.html$  /documentation/v1.1/configuration/stapel_image/$3.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/chart_configuration\.html$  /documentation/v1.1/reference/deploy_process/deploy_into_kubernetes.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/deploy_to_kubernetes\.html$  /documentation/v1.1/reference/deploy_process/deploy_into_kubernetes.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/minikube\.html$  /documentation/v1.1/reference/development_and_debug/setup_minikube.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/secrets\.html$  /documentation/v1.1/reference/deploy_process/working_with_secrets.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/track_kubernetes_resources\.html$  /documentation/v1.1/reference/deploy_process/differences_with_helm.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/authorization\.html$  /documentation/v1.1/reference/registry_authorization.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/cleaning\.html$  /documentation/v1.1/reference/cleaning_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/publish\.html$  /documentation/v1.1/reference/publish_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/tag\.html$  /documentation/v1.1/reference/publish_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/image_naming\.html$  /documentation/v1.1/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/toolbox/(.+)\.html$  /documentation/v1.1/reference/toolbox/$3.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?(/documentation)?/reference/registry/image_naming\.html$  /documentation/v1.1/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?(/documentation)?/reference/registry/push\.html$  /documentation/v1.1/reference/publish_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/cleanup_process\.html$  /documentation/v1.1/reference/cleaning_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/stage_introspection\.html$  /documentation/v1.1/reference/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/as_layers\.html$  /documentation/v1.1/reference/development_and_debug/as_layers.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/lint_and_render_chart\.html$  /documentation/v1.1/reference/development_and_debug/lint_and_render_chart.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/setup_minikube\.html$  /documentation/v1.1/reference/development_and_debug/setup_minikube.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/installing_minikube\.html$  /documentation/v1.1/reference/development_and_debug/setup_minikube.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/image_from_dockerfile\.html$  /documentation/v1.1/configuration/dockerfile_image.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/stages_and_images\.html$  /documentation/v1.1/internals/stages_and_storage.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/stages_and_images\.html$  /documentation/v1.1/internals/stages_and_storage.htmlpermanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/stage_introspection\.html$  /documentation/v1.1/advanced/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/development_and_debug/stage_introspection\.html$  /documentation/v1.1/advanced/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/assembly_process\.html$  /documentation/v1.1/configuration/stapel_image/assembly_instructions.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/stages\.html$  /documentation/v1.1/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/cli/toolbox/meta/get_helm_release\.html$  /documentation/v1.1/reference/cli/werf_helm_get_release.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/cli/management/helm/get_release\.html$  /documentation/v1.1/reference/cli/werf_helm_get_release.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/cli/$ /documentation/v1.1/reference/cli/overview.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/cli/toolbox/meta/get_helm_release\.html$  /documentation/v1.1/reference/cli/werf_helm_get_release.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/cli/(.+)$ /documentation/v1.1/cli/$3 permanent;
{{- end }}
