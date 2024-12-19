# SHOULD BE IN SYNC WITH github.com/werf/website/.helm/templates/_rewrites.tpl
{{- define "rewrites" }}

############################################
# Normalize urls
############################################

rewrite ^/js/(?<tail>.+)                                                               /assets/js/$tail        redirect;
rewrite ^/css/(?<tail>.+)                                                              /assets/css/$tail       redirect;
rewrite ^/images/(?<tail>.+)                                                           /assets/images/$tail    redirect;

rewrite ^/docs\.html$                                                                  /docs/                  redirect;
rewrite ^/docs/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)$                      /docs/$ver/             redirect;
rewrite ^/docs/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/index\.html$          /docs/$ver/             redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/docs\.html$                /docs/$ver/             redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/docs/?$                    /docs/$ver/             redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/docs/(?<tail>.+)           /docs/$ver/$tail        redirect;

rewrite ^/documentation\.html$                                                         /docs/                  redirect;
rewrite ^/documentation/?$                                                             /docs/                  redirect;
rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/?$           /docs/$ver/             redirect;
rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/index\.html$ /docs/$ver/             redirect;
rewrite ^/documentation/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/(?<tail>.+)  /docs/$ver/$tail        redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/documentation\.html$       /docs/$ver/             redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/documentation/?$           /docs/$ver/             redirect;
rewrite ^/(?<ver>v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/documentation/(?<tail>.+)  /docs/$ver/$tail        redirect;

rewrite ^/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/how_to/?$                           /docs/$ver/how_to/      redirect;
rewrite ^/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/how_to/(?<tail>.+)                  /docs/$ver/how_to/$tail redirect;

############################################
# Temporary versioned redirects
############################################

rewrite ^/docs/?$                                                                             /docs/v2/                                                       redirect;
rewrite ^/docs/(?!(v\d+(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/)(?:.+)                         /docs/v2/                                                       redirect;

rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/?$                             /docs/$ver/usage/project_configuration/overview.html                                                     redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/?$                       /docs/$ver/usage/project_configuration/overview.html                                                     redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/project_configuration/?$ /docs/$ver/usage/project_configuration/overview.html                                                     redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/build/?$                 /docs/$ver/usage/build/overview.html                                                                     redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/build/stapel/?$          /docs/$ver/usage/build/stapel/overview.html                                                              redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/deploy/?$                /docs/$ver/usage/deploy/overview.html                                                                    redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/distribute/?$            /docs/$ver/usage/distribute/overview.html                                                                redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/usage/cleanup/?$               /docs/$ver/usage/cleanup/cr_cleanup.html                                                                 redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/reference/?$                   /docs/$ver/reference/werf_yaml.html                                                                      redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/reference/cli/?$               /docs/$ver/reference/cli/overview.html                                                                   redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:\.\d+(?:[^/]+)?)?)?|latest)/resources/?$                   /docs/$ver/resources/cheat_sheet.html                                                                    redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:[^/]+)?)?|latest)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)$                                 /docs/$ver/reference/cli/werf_${tail1}_${tail2}                   redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:[^/]+)?)?|latest)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)/(?<tail3>[^/]+)$                 /docs/$ver/reference/cli/werf_${tail1}_${tail2}_${tail3}          redirect;
rewrite ^/docs/(?<ver>v2(?:\.\d+(?:[^/]+)?)?|latest)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)/(?<tail3>[^/]+)/(?<tail4>[^/]+)$ /docs/$ver/reference/cli/werf_${tail1}_${tail2}_${tail3}_${tail4} redirect;

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

rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/?$                                    /docs/$ver/index.html                                           redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/?$                              /docs/$ver/index.html                                           redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/project_configuration/?$        /docs/$ver/index.html                                           redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/configuration/?$                      /docs/$ver/configuration/introduction.html                      redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/configuration/stapel_image/?$         /docs/$ver/configuration/stapel_image/naming.html               redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/?$                          /docs/$ver/reference/stages_and_images.html                     redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/deploy_process/?$           /docs/$ver/reference/deploy_process/deploy_into_kubernetes.html redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/plugging_into_cicd/?$       /docs/$ver/reference/plugging_into_cicd/overview.html           redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/development_and_debug/?$    /docs/$ver/reference/development_and_debug/setup_minikube.html  redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/toolbox/?$                  /docs/$ver/reference/toolbox/slug.html                          redirect;

############################################
# Redirects for moved or deleted urls
############################################

rewrite ^/installation\.html$                                                                     /getting_started/                                      redirect;
rewrite ^/applications_guide_(?:ru|en)/?                                                          /guides.html                                           redirect;
rewrite ^/publications_ru\.html$                                                                  https://ru.werf.io/publications.html                   redirect;
rewrite ^/how_it_works\.html                                                                      /#how-it-works                                         redirect;
rewrite ^/introduction\.html$                                                                     /#how-it-works                                         redirect;

rewrite ^/guides/(?<lang>[^/]+)/400_ci_cd_workflow/030_gitlab_ci_cd/010_workflows\.html           /guides/$lang/400_ci_cd_workflow/030_gitlab_ci_cd.html redirect;
rewrite ^/guides/(?<lang>[^/]+)/400_ci_cd_workflow/030_gitlab_ci_cd/020_docker_executor\.html     /guides/$lang/400_ci_cd_workflow/030_gitlab_ci_cd.html redirect;
rewrite ^/guides/(?<lang>[^/]+)/400_ci_cd_workflow/030_gitlab_ci_cd/030_kubernetes_executor\.html /guides/$lang/400_ci_cd_workflow/030_gitlab_ci_cd.html redirect;

############################################
# v1.1/v1.2 redirects for moved or deleted urls
############################################

rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/quickstart\.html$                                                                  /docs/$ver/                                             redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/using_with_ci_cd_systems\.html$                                                    /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;

rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/supported_registry_implementations\.html$                                 /docs/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/buildah_mode\.html$                                                       /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/artifacts\.html$                              /docs/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/assembly_instructions\.html$                  /docs/$ver/usage/build/stapel/instructions.html         redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/base_image\.html$                             /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/docker_directive\.html$                       /docs/$ver/usage/build/stapel/dockerfile.html           redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/git_directive\.html$                          /docs/$ver/usage/build/stapel/git.html                  redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/import_directive\.html$                       /docs/$ver/usage/build/stapel/imports.html              redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/building_images_with_stapel/mount_directive\.html$                        /docs/$ver/usage/build/stapel/mounts.html               redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/bundles\.html$                                                            /docs/$ver/usage/distribute/bundles.html                redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/ci_cd_workflow_basics\.html$                                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/generic_ci_cd_integration\.html$                                    /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/github_actions\.html$                                               /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/gitlab_ci_cd\.html$                                                 /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/run_in_docker_container\.html$                     /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/run_in_kubernetes\.html$                           /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_docker_container\.html$                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_github_actions_with_docker_executor\.html$     /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_github_actions_with_kubernetes_executor\.html$ /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_docker_executor\.html$       /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_kubernetes_executor\.html$   /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/run_in_container/use_kubernetes\.html$                              /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview\.html$                         /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/werf_with_argocd/configure_ci_cd\.html$                             /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/ci_cd/werf_with_argocd/prepare_kubernetes_cluster\.html$                  /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/cleanup\.html$                                                            /docs/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/configuration/giterminism\.html$                                          /docs/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/configuration/organizing_configuration\.html$                             /docs/$ver/reference/werf_yaml_template_engine.html     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/configuration/supported_go_templates\.html$                               /docs/$ver/reference/werf_yaml_template_engine.html     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/giterminism\.html$                                                        /docs/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/chart\.html$                                           /docs/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/chart_dependencies\.html$                              /docs/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/giterminism\.html$                                     /docs/$ver/usage/project_configuration/giterminism.html redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/secrets\.html$                                         /docs/$ver/usage/deploy/values.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/templates\.html$                                       /docs/$ver/usage/deploy/templates.html                  redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/configuration/values\.html$                                          /docs/$ver/usage/deploy/values.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/annotating_and_labeling\.html$                        /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/deployment_order\.html$                               /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/external_dependencies\.html$                          /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/helm_hooks\.html$                                     /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/resources_adoption\.html$                             /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/deploy_process/steps\.html$                                          /docs/$ver/usage/deploy/deployment_order.html           redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/overview\.html$                                                      /docs/$ver/usage/deploy/overview.html                   redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/releases/manage_releases\.html$                                      /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/releases/naming\.html$                                               /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/releases/release\.html$                                              /docs/$ver/usage/deploy/releases.html                   redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/working_with_chart_dependencies\.html$                               /docs/$ver/usage/deploy/charts.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/helm/working_with_secrets\.html$                                          /docs/$ver/usage/deploy/values.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/storage_layouts\.html$                                                    /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/supported_container_registries\.html$                                     /docs/$ver/usage/cleanup/cr_cleanup.html                redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/advanced/synchronization\.html$                                                    /docs/$ver/usage/build/process.html                     redirect;

rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/build_process\.html$                                                     /docs/$ver/usage/build/process.html                     redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/development/stapel_image\.html$                                          /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/how_ci_cd_integration_works/general_overview\.html$                      /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/how_ci_cd_integration_works/github_actions\.html$                        /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/how_ci_cd_integration_works/gitlab_ci_cd\.html$                          /docs/$ver/usage/integration_with_ci_cd_systems.html    redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/integration_with_ssh_agent\.html$                                        /docs/$ver/usage/build/stapel/base.html                 redirect;
rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/internals/telemetry\.html$                                                         /docs/$ver/resources/telemetry.html                     redirect;

rewrite ^/docs/(?<ver>v1\.[12](?:\.\d+(?:[^/]+)?)?|latest)/reference/cheat_sheet\.html$                                                       /docs/$ver/resources/cheat_sheet.html                   redirect;

############################################
# v1.2 redirects for moved or deleted urls
############################################

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/cli/main/(?<tail>.+)                                                            /docs/$ver/reference/cli/werf_$tail                               redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)$                                 /docs/$ver/reference/cli/werf_${tail1}_${tail2}                   redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)/(?<tail3>[^/]+)$                 /docs/$ver/reference/cli/werf_${tail1}_${tail2}_${tail3}          redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/cli/management/(?<tail1>[^/]+)/(?<tail2>[^/]+)/(?<tail3>[^/]+)/(?<tail4>[^/]+)$ /docs/$ver/reference/cli/werf_${tail1}_${tail2}_${tail3}_${tail4} redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/cli/other/(?<tail>.+)                                                           /docs/$ver/reference/cli/werf_$tail                               redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/cli/toolbox/(?<tail>.+)                                                         /docs/$ver/reference/cli/werf_$tail                               redirect;

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/advanced_build/artifacts\.html$                                          /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/advanced_build/first_application\.html$                                  /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/advanced_build/mounts\.html$                                             /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/advanced_build/multi_images\.html$                                       /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/deploy_into_kubernetes\.html$                                            /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/generic_ci_cd_integration\.html$                                         /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/getting_started\.html$                                                   /getting_started/                                                 redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/github_ci_cd_integration\.html$                                          /docs/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/gitlab_ci_cd_integration\.html$                                          /docs/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/guides/installation\.html$                                                      /getting_started/                                                 redirect;

rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/build_process\.html$                                                  /docs/$ver/usage/build/process.html                               redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/ci_cd_workflows_overview\.html$                                       /docs/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/cleaning_process\.html$                                               /docs/$ver/usage/cleanup/cr_cleanup.html                          redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/deploy_process/deploy_into_kubernetes\.html$                          /docs/$ver/usage/deploy/overview.html                             redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/deploy_process/working_with_chart_dependencies\.html$                 /docs/$ver/usage/deploy/charts.html                               redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/development_and_debug/lint_and_render_chart\.html$                    /docs/$ver/                                                       redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/development_and_debug/stage_introspection\.html$                      /docs/$ver/usage/build/stapel/base.html                           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/plugging_into_cicd/gitlab_ci\.html$                                   /docs/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/plugging_into_cicd/overview\.html$                                    /docs/$ver/usage/integration_with_ci_cd_systems.html              redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/stages_and_images\.html$                                              /docs/$ver/usage/build/process.html                               redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/toolbox/slug\.html$                                                   /docs/$ver/                                                       redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/toolbox/ssh\.html$                                                    /docs/$ver/usage/build/stapel/base.html                           redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/working_with_docker_registries\.html$                                 /docs/$ver/usage/cleanup/cr_cleanup.html                          redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/whats_new_in_v1_2/changelog\.html$                                              /docs/$ver/resources/how_to_migrate_from_v1_1_to_v1_2.html        redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2\.html$                       /docs/$ver/resources/how_to_migrate_from_v1_1_to_v1_2.html        redirect;
rewrite ^/docs/(?<ver>v1\.2(?:\.\d+(?:[^/]+)?)?|latest)/reference/stages_and_storage\.html$                                             /docs/$ver/usage/build/process.html                               redirect;

############################################
# v1.1 redirects for moved or deleted urls
############################################

rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/stages_and_images\.html$                                              /docs/$ver/reference/build_process.html                                             redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/development_and_debug/stage_introspection\.html$                      /docs/$ver/configuration/stapel_image/base_image.html                               redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/build/process.html                                                        /docs/$ver/reference/build_process\.html                                            redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/integration_with_ci_cd_systems.html                                       /docs/$ver/reference/ci_cd_workflows_overview\.html                                 redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/deploy/overview.html                                                      /docs/$ver/reference/deploy_process/deploy_into_kubernetes\.html                    redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/deploy/charts.html                                                        /docs/$ver/reference/deploy_process/working_with_chart_dependencies\.html           redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/build/stapel/base.html                                                    /docs/$ver/reference/development_and_debug/stage_introspection\.html                redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/integration_with_ci_cd_systems.html                                       /docs/$ver/reference/plugging_into_cicd/gitlab_ci\.html                             redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/integration_with_ci_cd_systems.html                                       /docs/$ver/reference/plugging_into_cicd/overview\.html                              redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/build/process.html                                                        /docs/$ver/reference/stages_and_images\.html                                        redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/reference/cli/werf_slugify.html                                                 /docs/$ver/cli/toolbox/slugify.html                                                 redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/build/stapel/base.html                                                    /docs/$ver/reference/toolbox/ssh\.html                                              redirect;
rewrite ^/docs/(?<ver>v1\.1(?:\.\d+(?:[^/]+)?)?|latest)/usage/cleanup/cr_cleanup.html                                                   /docs/$ver/reference/working_with_docker_registries\.html                           redirect;

{{- end }}