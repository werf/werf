{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Render Kubernetes templates. This command will calculate digests and build (if needed) all images   
defined in the werf.yaml.

{{ header }} Syntax

```shell
werf render [options]
```

{{ header }} Environments

```shell
  $WERF_DEBUG_ANSIBLE_ARGS  Pass specified cli args to ansible ($ANSIBLE_ARGS)
  $WERF_SECRET_KEY          Use specified secret key to extract secrets for the deploy. Recommended 
                            way to set secret key in CI-system. 
                            
                            Secret key also can be defined in files:
                            * ~/.werf/global_secret_key (globally),
                            * .werf_secret_key (per project)
```

{{ header }} Options

```shell
      --add-annotation=[]
            Add annotation to deploying resources (can specify multiple).
            Format: annoName=annoValue.
            Also, can be specified with $WERF_ADD_ANNOTATION* (e.g.                                 
            $WERF_ADD_ANNOTATION_1=annoName1=annoValue1",                                           
            $WERF_ADD_ANNOTATION_2=annoName2=annoValue2")
      --add-label=[]
            Add label to deploying resources (can specify multiple).
            Format: labelName=labelValue.
            Also, can be specified with $WERF_ADD_LABEL* (e.g.                                      
            $WERF_ADD_LABEL_1=labelName1=labelValue1", $WERF_ADD_LABEL_2=labelName2=labelValue2")
      --atomic=false
            Enable auto rollback of the failed release to the previous deployed release version     
            when current deploy process have failed ($WERF_ATOMIC by default)
  -R, --auto-rollback=false
            Enable auto rollback of the failed release to the previous deployed release version     
            when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --dev=false
            Enable developer mode (default $WERF_DEV)
      --dir=''
            Use custom working directory (default $WERF_DIR or current directory)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and push images into the specified repo 
            and to pull base images
      --env=''
            Use specified environment (default $WERF_ENV)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --introspect-before-error=false
            Introspect failed stage in the clean state, before running all assembly instructions of 
            the stage
      --introspect-error=false
            Introspect failed stage in the state, right after running failed assembly instruction
      --introspect-stage=[]
            Introspect a specific stage. The option can be used multiple times to introspect        
            several stages.
            
            There are the following formats to use:
            * specify IMAGE_NAME/STAGE_NAME to introspect stage STAGE_NAME of either image or       
            artifact IMAGE_NAME
            * specify STAGE_NAME or */STAGE_NAME for the introspection of all existing stages with  
            name STAGE_NAME
            
            IMAGE_NAME is the name of an image or artifact described in werf.yaml, the nameless     
            image specified with ~.
            STAGE_NAME should be one of the following: from, beforeInstall, importsBeforeInstall,   
            gitArchive, install, importsAfterInstall, beforeSetup, importsBeforeSetup, setup,       
            importsAfterSetup, gitCache, gitLatestPatch, dockerInstructions, dockerfile
      --log-color-mode='auto'
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-project-dir=false
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-quiet=true
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --loose-giterminism=false
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info                                                                               
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_LOOSE_GITERMINISM)
      --namespace=''
            Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or         
            deploy.namespace custom template from werf.yaml or $WERF_NAMESPACE)
      --output=''
            Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by     
            default)
  -p, --parallel=true
            Run in parallel (default $WERF_PARALLEL)
      --parallel-tasks-limit=5
            Parallel tasks limit, set -1 to remove the limitation (default                          
            $WERF_PARALLEL_TASKS_LIMIT or 5)
      --release=''
            Use specified Helm release name (default [[ project ]]-[[ env ]] template or            
            deploy.helmRelease custom template from werf.yaml or $WERF_RELEASE)
      --repo=''
            Docker Repo to store stages (default $WERF_REPO)
      --repo-docker-hub-password=''
            Docker Hub password (default $WERF_REPO_DOCKER_HUB_PASSWORD)
      --repo-docker-hub-token=''
            Docker Hub token (default $WERF_REPO_DOCKER_HUB_TOKEN)
      --repo-docker-hub-username=''
            Docker Hub username (default $WERF_REPO_DOCKER_HUB_USERNAME)
      --repo-github-token=''
            GitHub token (default $WERF_REPO_GITHUB_TOKEN)
      --repo-harbor-password=''
            Harbor password (default $WERF_REPO_HARBOR_PASSWORD)
      --repo-harbor-username=''
            Harbor username (default $WERF_REPO_HARBOR_USERNAME)
      --repo-implementation=''
            Choose repo implementation.
            The following docker registry implementations are supported: ecr, acr, default,         
            dockerhub, gcr, github, gitlab, harbor, quay.
            Default $WERF_REPO_IMPLEMENTATION or auto mode (detect implementation by a registry).
      --repo-quay-token=''
            quay.io token (default $WERF_REPO_QUAY_TOKEN)
      --report-format='json'
            Report format: json or envfile (json or $WERF_REPORT_FORMAT by default)
            json:
            	{
            	  "Images": {
            		"<WERF_IMAGE_NAME>": {
            			"WerfImageName": "<WERF_IMAGE_NAME>",
            			"DockerRepo": "<REPO>",
            			"DockerTag": "<TAG>"
            			"DockerImageName": "<REPO>:<TAG>",
            			"DockerImageID": "<SHA256>",
            		},
            		...
            	  }
            	}
            envfile:
            	WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME=<REPO>:<TAG>
            	...
            <FORMATTED_WERF_IMAGE_NAME> is werf image name from werf.yaml modified according to the 
            following rules:
            - all characters are uppercase (app -> APP);
            - charset /- is replaced with _ (dev/app-frontend -> DEV_APP_FRONTEND)
      --report-path=''
            Report save path ($WERF_REPORT_PATH by default)
      --secondary-repo=[]
            Specify one or multiple secondary read-only repo with images that will be used as a     
            cache
      --secret-values=[]
            Specify helm secret values in a YAML file (can specify multiple).
            Also, can be defined with $WERF_SECRET_VALUES* (e.g.                                    
            $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml,                                  
            $WERF_SECRET_VALUES=.helm/secret_values_db.yaml)
      --set=[]
            Set helm values on the command line (can specify multiple or separate values with       
            commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET* (e.g. $WERF_SET_1=key1=val1, $WERF_SET_2=key2=val2)
      --set-file=[]
            Set values from respective files specified via the command line (can specify multiple   
            or separate values with commas: key1=path1,key2=path2).
            Also, can be defined with $WERF_SET_FILE* (e.g. $WERF_SET_FILE_1=key1=path1,            
            $WERF_SET_FILE_2=key2=val2)
      --set-string=[]
            Set STRING helm values on the command line (can specify multiple or separate values     
            with commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET_STRING* (e.g. $WERF_SET_STRING_1=key1=val1,         
            $WERF_SET_STRING_2=key2=val2)
  -Z, --skip-build=false
            Disable building of docker images, cached images in the repo should exist in the repo   
            if werf.yaml contains at least one image description (default $WERF_SKIP_BUILD)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --ssh-key=[]
            Use only specific ssh key(s).
            Can be specified with $WERF_SSH_KEY* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa",         
            $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa").
            Defaults to $WERF_SSH_KEY*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see             
            https://werf.io/documentation/reference/toolbox/ssh.html
  -S, --synchronization=''
            Address of synchronizer for multiple werf processes to work with a single repo.
            
            Default:
            * $WERF_SYNCHRONIZATION or
            * :local if --repo is not specified or
            * kubernetes://werf-synchronization if --repo is specified
            
            The same address should be specified for all werf processes that work with a single     
            repo. :local address allows execution of werf processes from a single host only
  -t, --timeout=0
            Resources tracking timeout in seconds
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --values=[]
            Specify helm values in a YAML file or a URL (can specify multiple).
            Also, can be defined with $WERF_VALUES* (e.g. $WERF_VALUES_ENV=.helm/values_test.yaml,  
            $WERF_VALUES_DB=.helm/values_db.yaml)
      --virtual-merge=false
            Enable virtual/ephemeral merge commit mode when building current application state      
            ($WERF_VIRTUAL_MERGE by default)
      --virtual-merge-from-commit=''
            Commit hash for virtual/ephemeral merge commit with new changes introduced in the pull  
            request ($WERF_VIRTUAL_MERGE_FROM_COMMIT by default)
      --virtual-merge-into-commit=''
            Commit hash for virtual/ephemeral merge commit which is base for changes introduced in  
            the pull request ($WERF_VIRTUAL_MERGE_INTO_COMMIT by default)
```

