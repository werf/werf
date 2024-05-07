# SHOULD BE IN SYNC WITH github.com/werf/website/.helm/templates/_rewrites.tpl
{{- define "rewrites" }}

############################################
# Normalize urls
############################################

rewrite ^/js/(?<tail>.+)                                                                     /assets/js/$tail                      redirect;
rewrite ^/css/(?<tail>.+)                                                                    /assets/css/$tail                     redirect;
rewrite ^/images/(?<tail>.+)                                                                 /assets/images/$tail                  redirect;

rewrite ^/documentation/latest/?$                                                            /documentation/                       redirect;
rewrite ^/documentation/latest/(?<tail>.+)                                                   /documentation/$tail                  redirect;
rewrite ^/latest/documentation/?$                                                            /documentation/                       redirect;
rewrite ^/latest/documentation/(?<tail>.+)                                                   /documentation/$tail                  redirect;

rewrite ^/documentation\.html$                                                               /documentation/                       redirect;
rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)$                         /documentation/$ver/                  redirect;
rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/index\.html$             /documentation/$ver/                  redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/documentation\.html$                   /documentation/$ver/                  redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/documentation/?$                       /documentation/$ver/                  redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/documentation/(?<tail>.+)              /documentation/$ver/$tail             redirect;

rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/configurator/?$          /documentation/$ver/configurator.html redirect;
rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/configurator/(?<tail>.+) /documentation/$ver/configurator.html redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/documentation/configurator\.html$      /documentation/$ver/configurator.html redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/documentation/configurator/?$          /documentation/$ver/configurator.html redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/documentation/configurator/(?<tail>.+) /documentation/$ver/configurator.html redirect;

rewrite ^/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/?$                                       /documentation/$ver/how_to/           redirect;
rewrite ^/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/(?<tail>.+)                              /documentation/$ver/how_to/$tail      redirect;

############################################
# Temporary versioned redirects
############################################

rewrite ^/documentation/?$                                                                       /documentation/v2/                                                       redirect;
rewrite ^/documentation/(?!v\d+(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?/)(?:.+)                           /documentation/v2/                                                       redirect;

rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/?$                             /documentation/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/?$                       /documentation/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/project_configuration/?$ /documentation/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/build/?$                 /documentation/$ver/usage/build/overview.html                            redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/build/stapel/?$          /documentation/$ver/usage/build/stapel/overview.html                     redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/deploy/?$                /documentation/$ver/usage/deploy/overview.html                           redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/distribute/?$            /documentation/$ver/usage/distribute/overview.html                       redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/usage/cleanup/?$               /documentation/$ver/usage/cleanup/cr_cleanup.html                        redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/reference/?$                   /documentation/$ver/reference/werf_yaml.html                             redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/reference/cli/?$               /documentation/$ver/reference/cli/overview.html                          redirect;
rewrite ^/documentation/(?<ver>v2(?:\.\d+(?:\.\d+(?:-[^/]+)?)?)?)/resources/?$                   /documentation/$ver/resources/cheat_sheet.html                           redirect;

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/?$                                    /documentation/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/?$                              /documentation/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/project_configuration/?$        /documentation/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/build/?$                        /documentation/$ver/usage/build/overview.html                            redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/build/stapel/?$                 /documentation/$ver/usage/build/stapel/overview.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/deploy/?$                       /documentation/$ver/usage/deploy/overview.html                           redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/distribute/?$                   /documentation/$ver/usage/distribute/overview.html                       redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/usage/cleanup/?$                      /documentation/$ver/usage/cleanup/cr_cleanup.html                        redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/?$                          /documentation/$ver/reference/werf_yaml.html                             redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/cli/?$                      /documentation/$ver/reference/cli/overview.html                          redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/resources/?$                          /documentation/$ver/resources/cheat_sheet.html                           redirect;

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/?$                                    /documentation/$ver/index.html                                           redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/?$                      /documentation/$ver/configuration/introduction.html                      redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/?$         /documentation/$ver/configuration/stapel_image/naming.html               redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/?$                          /documentation/$ver/reference/stages_and_images.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/deploy_process/?$           /documentation/$ver/reference/deploy_process/deploy_into_kubernetes.html redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/plugging_into_cicd/?$       /documentation/$ver/reference/plugging_into_cicd/overview.html           redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/development_and_debug/?$    /documentation/$ver/reference/development_and_debug/setup_minikube.html  redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/toolbox/?$                  /documentation/$ver/reference/toolbox/slug.html                          redirect;

############################################
# Redirects for moved or deleted urls
############################################

rewrite ^/installation\.html$            /getting_started/                    redirect;
rewrite ^/applications_guide_(?:ru|en)/? /guides.html                         redirect;
rewrite ^/publications_ru\.html$         https://ru.werf.io/publications.html redirect;
rewrite ^/how_it_works\.html             /#how-it-works                       redirect;
rewrite ^/introduction\.html$            /#how-it-works                       redirect;

rewrite ^/guides/(?<lang>[^/]+)/400_ci_cd_workflow/030_gitlab_ci_cd/010_workflows\.html           /guides/$lang/400_ci_cd_workflow/030_gitlab_ci_cd.html redirect;
rewrite ^/guides/(?<lang>[^/]+)/400_ci_cd_workflow/030_gitlab_ci_cd/020_docker_executor\.html     /guides/$lang/400_ci_cd_workflow/030_gitlab_ci_cd.html redirect;
rewrite ^/guides/(?<lang>[^/]+)/400_ci_cd_workflow/030_gitlab_ci_cd/030_kubernetes_executor\.html /guides/$lang/400_ci_cd_workflow/030_gitlab_ci_cd.html redirect;

############################################
# v1.1/v1.2 redirects for moved or deleted urls
############################################

rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/quickstart\.html$                                                                  /documentation/$ver/                                             redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/using_with_ci_cd_systems\.html$                                                    /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;

rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/supported_registry_implementations\.html$                                 /documentation/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/buildah_mode\.html$                                                       /documentation/$ver/usage/build/process.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/artifacts\.html$                              /documentation/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/assembly_instructions\.html$                  /documentation/$ver/usage/build/stapel/instructions.html         redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/base_image\.html$                             /documentation/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/docker_directive\.html$                       /documentation/$ver/usage/build/stapel/dockerfile.html           redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/git_directive\.html$                          /documentation/$ver/usage/build/stapel/git.html                  redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/import_directive\.html$                       /documentation/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/building_images_with_stapel/mount_directive\.html$                        /documentation/$ver/usage/build/stapel/mounts.html               redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/bundles\.html$                                                            /documentation/$ver/usage/distribute/bundles.html                redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/ci_cd_workflow_basics\.html$                                        /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/generic_ci_cd_integration\.html$                                    /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/github_actions\.html$                                               /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/gitlab_ci_cd\.html$                                                 /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/run_in_docker_container\.html$                     /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/run_in_kubernetes\.html$                           /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/use_docker_container\.html$                        /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/use_github_actions_with_docker_executor\.html$     /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/use_github_actions_with_kubernetes_executor\.html$ /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_docker_executor\.html$       /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_kubernetes_executor\.html$   /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/run_in_container/use_kubernetes\.html$                              /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview\.html$                         /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/werf_with_argocd/configure_ci_cd\.html$                             /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/ci_cd/werf_with_argocd/prepare_kubernetes_cluster\.html$                  /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/cleanup\.html$                                                            /documentation/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/configuration/giterminism\.html$                                          /documentation/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/configuration/organizing_configuration\.html$                             /documentation/$ver/reference/werf_yaml_template_engine.html     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/configuration/supported_go_templates\.html$                               /documentation/$ver/reference/werf_yaml_template_engine.html     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/development_and_debug/stage_introspection\.html$                          /documentation/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/giterminism\.html$                                                        /documentation/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/configuration/chart\.html$                                           /documentation/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/configuration/chart_dependencies\.html$                              /documentation/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/configuration/giterminism\.html$                                     /documentation/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/configuration/secrets\.html$                                         /documentation/$ver/usage/deploy/values.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/configuration/templates\.html$                                       /documentation/$ver/usage/deploy/templates.html                  redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/configuration/values\.html$                                          /documentation/$ver/usage/deploy/values.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/deploy_process/annotating_and_labeling\.html$                        /documentation/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/deploy_process/deployment_order\.html$                               /documentation/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/deploy_process/external_dependencies\.html$                          /documentation/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/deploy_process/helm_hooks\.html$                                     /documentation/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/deploy_process/resources_adoption\.html$                             /documentation/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/deploy_process/steps\.html$                                          /documentation/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/overview\.html$                                                      /documentation/$ver/usage/deploy/overview.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/releases/manage_releases\.html$                                      /documentation/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/releases/naming\.html$                                               /documentation/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/releases/release\.html$                                              /documentation/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/working_with_chart_dependencies\.html$                               /documentation/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/helm/working_with_secrets\.html$                                          /documentation/$ver/usage/deploy/values.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/storage_layouts\.html$                                                    /documentation/$ver/usage/build/process.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/supported_container_registries\.html$                                     /documentation/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/advanced/synchronization\.html$                                                    /documentation/$ver/usage/build/process.html                     redirect;

rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/build_process\.html$                                                     /documentation/$ver/usage/build/process.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/development/stapel_image\.html$                                          /documentation/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/how_ci_cd_integration_works/general_overview\.html$                      /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/how_ci_cd_integration_works/github_actions\.html$                        /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/how_ci_cd_integration_works/gitlab_ci_cd\.html$                          /documentation/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/integration_with_ssh_agent\.html$                                        /documentation/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/stages_and_storage\.html$                                                /documentation/$ver/usage/build/process.html                     redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/internals/telemetry\.html$                                                         /documentation/$ver/resources/telemetry.html                     redirect;

rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/reference/build/artifact\.html$                                                    /documentation/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/documentation/(?<ver>v1\.[12](?:\.\d+(?:-[^/]+)?)?)/reference/cheat_sheet\.html$                                                       /documentation/$ver/resources/cheat_sheet.html                   redirect;

############################################
# v1.2 redirects for moved or deleted urls
############################################

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/cleanup\.html$                                                    /documentation/$ver/reference/werf_yaml.html#cleanup                       redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/deploy_into_kubernetes\.html$                                     /documentation/$ver/reference/werf_yaml.html#deploy                        redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/dockerfile_image\.html$                                           /documentation/$ver/reference/werf_yaml.html                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/introduction\.html$                                               /documentation/$ver/reference/werf_yaml.html                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_artifact\.html$                                            /documentation/$ver/usage/build/stapel/imports.html                        redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/image_directives\.html$                              /documentation/$ver/reference/werf_yaml.html                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/naming\.html$                                        /documentation/$ver/reference/werf_yaml.html#image-section                 redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/(?<tail>.+)                                          /documentation/$ver/advanced/building_images_with_stapel/$tail             redirect;

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/cli/main/(?<tail>.+)                                                            /documentation/$ver/reference/cli/werf_$tail                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)$                                 /documentation/$ver/reference/cli/werf_${tail1}_${tail2}                   redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)/(?<tail3>[^/]+)$                 /documentation/$ver/reference/cli/werf_${tail1}_${tail2}_${tail3}          redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)/(?<tail3>[^/]+)/(?<tail4>[^/]+)$ /documentation/$ver/reference/cli/werf_${tail1}_${tail2}_${tail3}_${tail4} redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/cli/other/(?<tail>.+)                                                           /documentation/$ver/reference/cli/werf_$tail                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/cli/toolbox/(?<tail>.+)                                                         /documentation/$ver/reference/cli/werf_$tail                               redirect;

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/development/stapel\.html$                                                       /documentation/$ver/usage/build/stapel/base.html                           redirect;

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/advanced_build/artifacts\.html$                                          /guides.html                                                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/advanced_build/first_application\.html$                                  /guides.html                                                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/advanced_build/mounts\.html$                                             /guides.html                                                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/advanced_build/multi_images\.html$                                       /guides.html                                                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/deploy_into_kubernetes\.html$                                            /documentation/$ver/quickstart.html                                        redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/generic_ci_cd_integration\.html$                                         /documentation/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/getting_started\.html$                                                   /documentation/$ver/quickstart.html                                        redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/github_ci_cd_integration\.html$                                          /documentation/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/gitlab_ci_cd_integration\.html$                                          /documentation/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/guides/installation\.html$                                                      /documentation/$ver/                                                       redirect;

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/build_process\.html$                                                  /documentation/$ver/usage/build/process.html                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/ci_cd_workflows_overview\.html$                                       /documentation/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/cleaning_process\.html$                                               /documentation/$ver/usage/cleanup/cr_cleanup.html                          redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/deploy_process/deploy_into_kubernetes\.html$                          /documentation/$ver/usage/deploy/overview.html                             redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/deploy_process/working_with_chart_dependencies\.html$                 /documentation/$ver/usage/deploy/charts.html                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/development_and_debug/lint_and_render_chart\.html$                    /documentation/$ver/                                                       redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/development_and_debug/stage_introspection\.html$                      /documentation/$ver/usage/build/stapel/base.html                           redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/plugging_into_cicd/gitlab_ci\.html$                                   /documentation/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/plugging_into_cicd/overview\.html$                                    /documentation/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/stages_and_images\.html$                                              /documentation/$ver/usage/build/process.html                               redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/toolbox/slug\.html$                                                   /documentation/$ver/                                                       redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/toolbox/ssh\.html$                                                    /documentation/$ver/usage/build/stapel/base.html                           redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/reference/working_with_docker_registries\.html$                                 /documentation/$ver/usage/cleanup/cr_cleanup.html                          redirect;

rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/whats_new_in_v1_2/changelog\.html$                                              /documentation/$ver/resources/how_to_migrate_from_v1_1_to_v1_2.html        redirect;
rewrite ^/documentation/(?<ver>v1\.2(?:\.\d+(?:-[^/]+)?)?)/whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2\.html$                       /documentation/$ver/resources/how_to_migrate_from_v1_1_to_v1_2.html        redirect;

############################################
# v1.1 redirects for moved or deleted urls
############################################

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/quickstart\.html$                                          /documentation/$ver/guides/getting_started.html                                redirect;

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/?$                                                  /documentation/$ver/guides/                                                    redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/mounts\.html$                                       /documentation/$ver/guides/advanced_build/mounts.html                          redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/multi_images\.html$                                 /documentation/$ver/guides/advanced_build/multi_images.html                    redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/artifacts\.html$                                    /documentation/$ver/guides/advanced_build/artifacts.html                       redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/how_to/(?<tail>.+)                                         /documentation/$ver/guides/$tail                                               redirect;

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/guides/guides/unsupported_ci_cd_integration\.html$         /documentation/$ver/guides/generic_ci_cd_integration.html                      redirect;

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/cli/?$                                                     /documentation/$ver/reference/cli/                                             redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/cli/management/helm/get_release\.html$                     /documentation/$ver/reference/cli/werf_helm_get_release.html                   redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/cli/toolbox/meta/get_helm_release\.html$                   /documentation/$ver/reference/cli/werf_helm_get_release.html                   redirect;

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/assembly_process\.html$         /documentation/$ver/configuration/stapel_image/assembly_instructions.html      redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/image_from_dockerfile\.html$    /documentation/$ver/configuration/dockerfile_image.html                        redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/stage_introspection\.html$      /documentation/$ver/advanced/development_and_debug/stage_introspection.html    redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/stages\.html$                   /documentation/$ver/reference/stages_and_images.html                           redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/configuration/stapel_image/stages_and_images\.html$        /documentation/$ver/internals/stages_and_storage.html                          redirect;

rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/cleanup_process\.html$                           /documentation/$ver/reference/cleaning_process.html                            redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/config\.html$                                    /documentation/$ver/configuration/introduction.html                            redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/stages_and_images\.html$                         /documentation/$ver/internals/stages_and_storage.html                          redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/build/as_layers\.html$                           /documentation/$ver/reference/development_and_debug/as_layers.html             redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/build/stage_introspection\.html$                 /documentation/$ver/reference/development_and_debug/stage_introspection.html   redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/build/(?<tail>.+)                                /documentation/$ver/configuration/stapel_image/$tail                           redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/deploy/chart_configuration\.html$                /documentation/$ver/reference/deploy_process/deploy_into_kubernetes.html       redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/deploy/deploy_to_kubernetes\.html$               /documentation/$ver/reference/deploy_process/deploy_into_kubernetes.html       redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/deploy/minikube\.html$                           /documentation/$ver/reference/development_and_debug/setup_minikube.html        redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/deploy/secrets\.html$                            /documentation/$ver/reference/deploy_process/working_with_secrets.html         redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/deploy/track_kubernetes_resources\.html$         /documentation/$ver/reference/deploy_process/differences_with_helm.html        redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/development_and_debug/stage_introspection\.html$ /documentation/$ver/advanced/development_and_debug/stage_introspection.html    redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/local_development/as_layers\.html$               /documentation/$ver/reference/development_and_debug/as_layers.html             redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/local_development/installing_minikube\.html$     /documentation/$ver/reference/development_and_debug/setup_minikube.html        redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/local_development/lint_and_render_chart\.html$   /documentation/$ver/reference/development_and_debug/lint_and_render_chart.html redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/local_development/setup_minikube\.html$          /documentation/$ver/reference/development_and_debug/setup_minikube.html        redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/local_development/stage_introspection\.html$     /documentation/$ver/reference/development_and_debug/stage_introspection.html   redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/registry/authorization\.html$                    /documentation/$ver/reference/registry_authorization.html                      redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/registry/cleaning\.html$                         /documentation/$ver/reference/cleaning_process.html                            redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/registry/image_naming\.html$                     /documentation/$ver/reference/stages_and_images.html                           redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/registry/publish\.html$                          /documentation/$ver/reference/publish_process.html                             redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/registry/push\.html$                             /documentation/$ver/reference/publish_process.html                             redirect;
rewrite ^/documentation/(?<ver>v1\.1(?:\.\d+(?:-[^/]+)?)?)/reference/registry/tag\.html$                              /documentation/$ver/reference/publish_process.html                             redirect;

{{- end }}
