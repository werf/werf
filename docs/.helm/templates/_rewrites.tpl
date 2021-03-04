{{- define "rewrites" }}
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)+/?$  $1/documentation/ permanent;

rewrite ^/applications_guide(_ru|_en)+(/.+)$   /guides$2 permanent;
rewrite ^/applications_guide(/.+)$   /guides$1 permanent;

rewrite ^(/([^/]+/)?)+documentation/advanced/configuration/organizing_configuration\.html $1documentation/reference/werf_yaml_template_engine.html permanent;

rewrite ^/documentation/advanced/configuration/supported_go_templates\.html$   /documentation/reference/werf_yaml_template_engine.html permanent;
rewrite ^/documentation/advanced/configuration/giterminism\.html$   /documentation/advanced/giterminism.html permanent;

rewrite ^(/v1\.[01]+(\-[a-z]+)?)+/documentation/quickstart\.html$            $1/documentation/guides/getting_started.html permanent;

rewrite ^/documentation/configuration/introduction\.html$                   /documentation/reference/werf_yaml.html permanent;
rewrite ^/documentation/configuration/dockerfile_image\.html$               /documentation/reference/werf_yaml.html permanent;
rewrite ^/documentation/configuration/stapel_image/naming\.html$            /documentation/reference/werf_yaml.html#image-section permanent;
rewrite ^/documentation/configuration/stapel_image/(.+)\.html$              /documentation/advanced/building_images_with_stapel/$1.html permanent;
rewrite ^/documentation/configuration/stapel_image/image_directives\.html$  /documentation/reference/werf_yaml.html permanent;
rewrite ^/documentation/configuration/stapel_artifact\.html$                /documentation/advanced/building_images_with_stapel/artifact.html permanent;
rewrite ^/documentation/configuration/deploy_into_kubernetes\.html$         /documentation/reference/werf_yaml.html#deploy  permanent;
rewrite ^/documentation/configuration/cleanup\.html$                        /documentation/reference/werf_yaml.html#cleanup  permanent;

rewrite ^/documentation/reference/build_process\.html$                                        /documentation/internals/build_process.html permanent;
rewrite ^/documentation/reference/stages_and_images\.html$                                    /documentation/internals/stages_and_storage.html permanent;
rewrite ^/documentation/reference/deploy_process/deploy_into_kubernetes\.html$                /documentation/advanced/helm/overview.html permanent;
rewrite ^(/v1\.[01]+(\-[a-z]+)?)?/documentation/advanced/helm/basics\.html$                    $1/documentation/advanced/helm/overview.html permanent;
rewrite ^(/v1\.[01]+(\-[a-z]+)?)?/documentation/reference/deploy_process/working_with_secrets\.html$                  /documentation/advanced/helm/configuration/secrets.html permanent;
rewrite ^/documentation/advanced/helm/working_with_secrets\.html$                             /documentation/advanced/helm/configuration/secrets.html permanent;

rewrite ^/documentation/reference/deploy_process/working_with_chart_dependencies\.html$       /documentation/advanced/helm/configuration/chart_dependencies.html permanent;
rewrite ^/documentation/advanced/helm/working_with_chart_dependencies\.html$                  /documentation/advanced/helm/configuration/chart_dependencies.html permanent;

rewrite ^/documentation/reference/cleaning_process\.html$                                     /documentation/advanced/cleanup.html permanent;
rewrite ^/documentation/reference/working_with_docker_registries\.html$                       /documentation/advanced/supported_registry_implementations.html permanent;
rewrite ^/documentation/reference/ci_cd_workflows_overview\.html$                             /documentation/advanced/ci_cd/ci_cd_workflow_basics.html permanent;
rewrite ^/documentation/reference/plugging_into_cicd/overview\.html$                          /documentation/internals/how_ci_cd_integration_works/general_overview.html permanent;
rewrite ^/documentation/reference/plugging_into_cicd/gitlab_ci\.html$                         /documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html permanent;
rewrite ^/documentation/reference/development_and_debug/stage_introspection\.html$            /documentation/advanced/development_and_debug/stage_introspection.html permanent;

rewrite ^/documentation/reference/development_and_debug/lint_and_render_chart\.html$          /documentation/advanced/development_and_debug/lint_and_render_chart.html permanent;
rewrite ^/documentation/reference/toolbox/slug\.html$                                         /documentation/internals/names_slug_algorithm.html permanent;
rewrite ^/documentation/reference/toolbox/ssh\.html$                                          /documentation/internals/integration_with_ssh_agent.html permanent;
rewrite ^/documentation/cli/(main|toolbox|other)/([^/]+)\.html$                            /documentation/reference/cli/werf_$2.html permanent;
rewrite ^/documentation/cli/management/([^/]+)/([^/]+)\.html$                        /documentation/reference/cli/werf_$1_$2.html permanent;
rewrite ^/documentation/cli/management/([^/]+)/([^/]+)/([^/]+)\.html$                /documentation/reference/cli/werf_$1_$2_$3.html permanent;
rewrite ^/documentation/cli/management/([^/]+)/([^/]+)/([^/]+)/([^/]+)\.html$        /documentation/reference/cli/werf_$1_$2_$3_$4.html permanent;
rewrite ^/documentation/development/stapel\.html$                                    /documentation/internals/development/stapel_image.html permanent;
rewrite ^/documentation/guides/installation\.html$                                   /installation.html permanent;
rewrite ^(/v1\.[^01]+(\-[a-z]+)?)+/documentation/guides/(getting_started|deploy_into_kubernetes)+\.html$       $1/documentation/quickstart.html permanent;
rewrite ^/documentation/guides/(getting_started|deploy_into_kubernetes)+\.html$      /documentation/quickstart.html permanent;
rewrite ^/documentation/guides/generic_ci_cd_integration\.html$                      /documentation/advanced/ci_cd/generic_ci_cd_integration.html permanent;
rewrite ^/documentation/guides/gitlab_ci_cd_integration\.html$                       /documentation/advanced/ci_cd/gitlab_ci_cd.html permanent;
rewrite ^/documentation/guides/github_ci_cd_integration\.html$                       /documentation/advanced/ci_cd/github_actions.html permanent;
rewrite ^/documentation/guides/advanced_build/(first_application|multi_images|mounts|artifacts)+\.html$    /documentation/guides.html permanent;

rewrite ^(/v1.1(\-(alpha|beta|ea|stable)+)?)+/documentation/guides/unsupported_ci_cd_integration\.html$ $1/documentation/guides/generic_ci_cd_integration.html permanent;
rewrite ^(/v1.0(\-(alpha|beta|ea|stable|rock-solid)+)?)+/documentation/guides/generic_ci_cd_integration\.html$ $1/documentation/advanced/ci_cd/generic_ci_cd_integration.html permanent;
rewrite ^(/v1.0(\-(alpha|beta|ea|stable|rock-solid)+)?)+/documentation/guides/unsupported_ci_cd_integration\.html$ $1/documentation/advanced/ci_cd/generic_ci_cd_integration.html permanent;

rewrite ^/publications_ru\.html$  https://ru.werf.io/publications.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation\.html$  $1/documentation/ permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/?$  /installation.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/guides/installation\.html$  /installation.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/mounts\.html$  $1/documentation/guides/advanced_build/mounts.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/multi_images\.html$  $1/documentation/guides/advanced_build/multi_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/artifacts\.html$  $1/documentation/guides/advanced_build/artifacts.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/how_to/(.+)  $1/documentation/guides/$3 permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/?$  $1/documentation/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/config\.html$  $1/documentation/configuration/introduction.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/artifact\.html$  /documentation/advanced/building_images_with_stapel/artifacts.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/as_layers\.html$  $1/documentation/reference/development_and_debug/as_layers.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/stage_introspection\.html$  $1/documentation/reference/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/build/(.+)\.html$  $1/documentation/configuration/stapel_image/$3.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/chart_configuration\.html$  $1/documentation/reference/deploy_process/deploy_into_kubernetes.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/deploy_to_kubernetes\.html$  $1/documentation/reference/deploy_process/deploy_into_kubernetes.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/minikube\.html$  $1/documentation/reference/development_and_debug/setup_minikube.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/secrets\.html$  $1/documentation/reference/deploy_process/working_with_secrets.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/deploy/track_kubernetes_resources\.html$  $1/documentation/reference/deploy_process/differences_with_helm.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/authorization\.html$  $1/documentation/reference/registry_authorization.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/cleaning\.html$  $1/documentation/reference/cleaning_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/publish\.html$  $1/documentation/reference/publish_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/tag\.html$  $1/documentation/reference/publish_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/registry/image_naming\.html$  $1/documentation/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/reference/toolbox/(.+)\.html$  $1/documentation/reference/toolbox/$3.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?(/documentation)?/reference/registry/image_naming\.html$  $1/documentation/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?(/documentation)?/reference/registry/push\.html$  $1/documentation/reference/publish_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/cleanup_process\.html$  $1/documentation/reference/cleaning_process.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/stage_introspection\.html$  $1/documentation/reference/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/as_layers\.html$  $1/documentation/reference/development_and_debug/as_layers.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/lint_and_render_chart\.html$  $1/documentation/reference/development_and_debug/lint_and_render_chart.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/setup_minikube\.html$  $1/documentation/reference/development_and_debug/setup_minikube.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/local_development/installing_minikube\.html$  $1/documentation/reference/development_and_debug/setup_minikube.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/image_from_dockerfile\.html$  $1/documentation/configuration/dockerfile_image.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/stages_and_images\.html$  $1/documentation/internals/stages_and_storage.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/stages_and_images\.html$  $1/documentation/internals/stages_and_storage.htmlpermanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/stage_introspection\.html$  $1/documentation/advanced/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/reference/development_and_debug/stage_introspection\.html$  $1/documentation/advanced/development_and_debug/stage_introspection.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/assembly_process\.html$  $1/documentation/configuration/stapel_image/assembly_instructions.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/configuration/stapel_image/stages\.html$  $1/documentation/reference/stages_and_images.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/cli/toolbox/meta/get_helm_release\.html$  $1/documentation/reference/cli/werf_helm_get_release.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/documentation/cli/management/helm/get_release\.html$  $1/documentation/reference/cli/werf_helm_get_release.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/cli/$ $1/documentation/reference/cli/overview.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/cli/toolbox/meta/get_helm_release\.html$  $1/documentation/reference/cli/werf_helm_get_release.html permanent;
rewrite ^(/v[\d]+\.[\d]+(\-[a-z]+)?)?/cli/(.+)$ $1/documentation/cli/$3 permanent;
{{- end }}
