# SHOULD BE IN SYNC WITH github.com/werf/website/.helm/templates/_rewrites.tpl
{{- define "rewrites" }}

############################################
# Normalize urls
############################################

rewrite ^/js/(?<tail>.+)                                                               /assets/js/$tail        redirect;
rewrite ^/css/(?<tail>.+)                                                              /assets/css/$tail       redirect;
rewrite ^/images/(?<tail>.+)                                                           /assets/images/$tail    redirect;

rewrite ^/docs\.html$                                                                  /docs/                  redirect;
rewrite ^/docs/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)$                 /docs/$ver/             redirect;
rewrite ^/docs/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/index\.html$     /docs/$ver/             redirect;
rewrite ^/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/docs\.html$           /docs/$ver/             redirect;
rewrite ^/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/docs/?$               /docs/$ver/             redirect;
rewrite ^/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/docs/(?<tail>.+)      /docs/$ver/$tail        redirect;

rewrite ^/documentation\.html$                                                         /docs/                  redirect;
rewrite ^/documentation/?$                                                             /docs/                  redirect;
rewrite ^/documentation/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/?$      /docs/$ver/             redirect;
rewrite ^/documentation/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/index\.html$ /docs/$ver/             redirect;
rewrite ^/documentation/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/(?<tail>.+) /docs/$ver/$tail        redirect;
rewrite ^/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/documentation\.html$  /docs/$ver/             redirect;
rewrite ^/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/documentation/?$      /docs/$ver/             redirect;
rewrite ^/(?<ver>latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/documentation/(?<tail>.+) /docs/$ver/$tail        redirect;

############################################
# Temporary versioned redirects
############################################

rewrite ^/docs/?$                                                                             /docs/v2/                                                       redirect;
rewrite ^/docs/(?!(latest|pr-[^/]+|v2(?:\.[^/]+)?|v1\.2(?:\.[^/]+)?)/)(?:.+)                    /docs/v2/                                                       redirect;

rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/?$                             /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/?$                       /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/project_configuration/?$ /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/build/?$                 /docs/$ver/usage/build/overview.html                            redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/build/stapel/?$          /docs/$ver/usage/build/stapel/overview.html                     redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/deploy/?$                /docs/$ver/usage/deploy/overview.html                           redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/distribute/?$            /docs/$ver/usage/distribute/overview.html                       redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/cleanup/?$               /docs/$ver/usage/cleanup/cr_cleanup.html                        redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/reference/?$                   /docs/$ver/reference/werf_yaml.html                             redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/reference/cli/?$               /docs/$ver/reference/cli/overview.html                          redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/resources/?$                   /docs/$ver/resources/cheat_sheet.html                           redirect;

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/?$                                    /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/?$                              /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/project_configuration/?$        /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/build/?$                        /docs/$ver/usage/build/overview.html                            redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/build/stapel/?$                 /docs/$ver/usage/build/stapel/overview.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/deploy/?$                       /docs/$ver/usage/deploy/overview.html                           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/distribute/?$                   /docs/$ver/usage/distribute/overview.html                       redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/usage/cleanup/?$                      /docs/$ver/usage/cleanup/cr_cleanup.html                        redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/?$                          /docs/$ver/reference/werf_yaml.html                             redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/cli/?$                      /docs/$ver/reference/cli/overview.html                          redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/resources/?$                          /docs/$ver/resources/cheat_sheet.html                           redirect;


############################################
# Redirects for moved or deleted urls
############################################

rewrite ^/installation\.html$                                                                     /getting_started/                                      redirect;
rewrite ^/applications_guide_(?:ru|en)/?                                                          /guides.html                                           redirect;
rewrite ^/publications_ru\.html$                                                                  https://ru.werf.io/publications.html                   redirect;
rewrite ^/how_it_works\.html                                                                      /#how-it-works                                         redirect;
rewrite ^/introduction\.html$                                                                     /#how-it-works                                         redirect;

############################################
# v1.2 redirects for moved or deleted urls
############################################

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/quickstart\.html$                                                                      /docs/$ver/                                             redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/using_with_ci_cd_systems\.html$                                                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/supported_registry_implementations\.html$                                     /docs/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/buildah_mode\.html$                                                           /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/artifacts\.html$                                  /docs/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/assembly_instructions\.html$                  /docs/$ver/usage/build/stapel/instructions.html         redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/base_image\.html$                             /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/docker_directive\.html$                       /docs/$ver/usage/build/stapel/dockerfile.html           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/git_directive\.html$                          /docs/$ver/usage/build/stapel/git.html                  redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/import_directive\.html$                       /docs/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/mount_directive\.html$                        /docs/$ver/usage/build/stapel/mounts.html               redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/bundles\.html$                                                            /docs/$ver/usage/distribute/bundles.html                redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/ci_cd_workflow_basics\.html$                                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/generic_ci_cd_integration\.html$                                    /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/github_actions\.html$                                               /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/gitlab_ci_cd\.html$                                                 /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/run_in_docker_container\.html$                     /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/run_in_kubernetes\.html$                           /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_docker_container\.html$                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_github_actions_with_docker_executor\.html$     /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_github_actions_with_kubernetes_executor\.html$ /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_docker_executor\.html$       /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_kubernetes_executor\.html$   /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_kubernetes\.html$                              /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview\.html$                         /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/werf_with_argocd/configure_ci_cd\.html$                             /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/werf_with_argocd/prepare_kubernetes_cluster\.html$                  /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/cleanup\.html$                                                            /docs/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/configuration/giterminism\.html$                                          /docs/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/configuration/organizing_configuration\.html$                             /docs/$ver/usage/project_configuration/werf_yaml_template_engine.html     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/configuration/supported_go_templates\.html$                               /docs/$ver/usage/project_configuration/werf_yaml_template_engine.html     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/development_and_debug/stage_introspection\.html$                          /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/giterminism\.html$                                                        /docs/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/chart\.html$                                           /docs/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/chart_dependencies\.html$                              /docs/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/giterminism\.html$                                     /docs/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/secrets\.html$                                         /docs/$ver/usage/deploy/values.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/templates\.html$                                       /docs/$ver/usage/deploy/templates.html                  redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/values\.html$                                          /docs/$ver/usage/deploy/values.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/annotating_and_labeling\.html$                        /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/deployment_order\.html$                               /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/external_dependencies\.html$                          /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/helm_hooks\.html$                                     /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/resources_adoption\.html$                             /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/steps\.html$                                          /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/overview\.html$                                                      /docs/$ver/usage/deploy/overview.html                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/releases/manage_releases\.html$                                      /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/releases/naming\.html$                                               /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/releases/release\.html$                                              /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/working_with_chart_dependencies\.html$                               /docs/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/working_with_secrets\.html$                                          /docs/$ver/usage/deploy/values.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/storage_layouts\.html$                                                    /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/supported_container_registries\.html$                                     /docs/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/advanced/synchronization\.html$                                                    /docs/$ver/usage/build/process.html                     redirect;

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/build_process\.html$                                                     /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/development/stapel_image\.html$                                          /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/how_ci_cd_integration_works/general_overview\.html$                      /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/how_ci_cd_integration_works/github_actions\.html$                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/how_ci_cd_integration_works/gitlab_ci_cd\.html$                          /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/integration_with_ssh_agent\.html$                                        /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/stages_and_storage\.html$                                                /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/internals/telemetry\.html$                                                         /docs/$ver/resources/telemetry.html                     redirect;

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/build/artifact\.html$                                                    /docs/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/cheat_sheet\.html$                                                       /docs/$ver/resources/cheat_sheet.html                   redirect;

{{- end }}
