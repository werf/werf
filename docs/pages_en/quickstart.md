---
title: Quickstart
permalink: quickstart.html
description: Deploy your first application with werf
---

In this article we will show you how to set up the deployment of an [example application](https://github.com/werf/quickstart-application) (a cool voting app in our case) using werf. It is better to start with a [short introduction](/how_it_works.html) first if you haven't read it yet.

## Prepare your host

 1. Install [dependencies](/installation.html#install-dependencies) (Docker and Git).
 2. Install [trdl and werf](/installation.html#install-werf).

Make sure you have `werf` command available in your shell before proceeding to the next step:
   
```
werf version
```

## Prepare your Kubernetes and container registry

You should have access to the Kubernetes cluster and be able to push images to your container registry. The container registry should also be accessible from the Kubernetes cluster to pull images.

If your Kubernetes and container registry are running already:

 1. Perform the standard docker login procedure into your container registry from your host.
 2. Make sure your Kubernetes cluster is accessible from your host (if the `kubectl` tool is already set up and running, then `werf` will also work just fine).

<br>

Or use one of the following instructions to set up the local Kubernetes cluster and container registry in your OS:

<div class="details">
<a href="javascript:void(0)" class="details__summary">Windows — minikube</a>
<div class="details__content" markdown="1">
1. Install [minikube](https://github.com/kubernetes/minikube#installation).
2. Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/).
3. Start minikube:

   {% raw %}
   ```shell
   minikube start --driver=docker --insecure-registry registry.example.com:80
   ```
   {% endraw %}
    
   **IMPORTANT** Param `--insecure-registry` allows usage of Container Registry without TLS. TLS in our case dropped for simplicity.

4. Install NGINX Ingress Controller:

   {% raw %}
   ```shell
   minikube addons enable ingress
   ```
   {% endraw %}

5. Install Container Registry to store images:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}

   Create Ingress to access Container Registry. 
   
       {% raw %}
   ```shell
   @"
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: registry
     namespace: kube-system
     annotations:
       nginx.ingress.kubernetes.io/proxy-body-size: "0"
   spec:
     rules:
     - host: registry.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: registry
               port:
                 number: 80
   "@ | kubectl apply -f -  
   ```
   {% endraw %}

6. Allow usage of Container Registry without TLS for Docker:

   Using menu Docker Desktop -> Settings -> Docker Engine add following configuration key:

   ```json
   {
      "insecure-registries": ["registry.example.com:80"]
   }
   ```

   Restart Docker Desktop using right button menu of the tray Docker Desktop icon.

   Then start minikube again:

   ```shell
   minikube start --driver=docker --insecure-registry registry.example.com:80
   ```

7. Allow usage of Container Registry without TLS for werf:

   Set `WERF_INSECURE_REGISTRY=1` environment variable in the terminal where werf would run.

   For cmd.exe:

   ```
   set WERF_INSECURE_REGISTRY=1
   ```

   For bash:

   ```
   export WERF_INSECURE_REGISTRY=1
   ```

   For PowerShell:

   ```
   $Env:WERF_INSECURE_REGISTRY = "1"
   ```

8. We are going to use `vote.quickstart-application.example.com` and `result.quickstart-application.example.com` domains to access application and `registry.example.com` domain to access Container Registry.

   Let's update hosts file. Get minikube IP-address:

   ```shell
   minikube ip
   ```

   Using IP-address above append line to the end of the file `C:\Windows\System32\drivers\etc\hosts`:
   
   ```
   <IP-address minikube>    vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com
   ```

   Should look like:
   ```
   192.168.99.99          vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com
   ```

9. Let's also add `registry.example.com` domain to the minikube node:

   ```shell
   minikube ssh -- "echo $(minikube ip) registry.example.com | sudo tee -a /etc/hosts"
   ```
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS — minikube</a>
<div class="details__content" markdown="1">
1. Install [minikube](https://github.com/kubernetes/minikube#installation).
2. Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/).
3. Start minikube:

   {% raw %}
   ```shell
   minikube start --vm=true --insecure-registry registry.example.com:80
   ```
   {% endraw %}
    
   **IMPORTANT** Param `--insecure-registry` allows usage of Container Registry without TLS. TLS in our case dropped for simplicity.

4. Install NGINX Ingress Controller:

   {% raw %}
   ```shell
   minikube addons enable ingress
   ```
   {% endraw %}

5. Install Container Registry to store images:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}
    
   Create Ingress to access Container Registry:
 
   {% raw %}
   ```shell
   kubectl apply -f - << EOF
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: registry
     namespace: kube-system
     annotations:
       nginx.ingress.kubernetes.io/proxy-body-size: "0"
   spec:
     rules:
     - host: registry.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: registry
               port:
                 number: 80
   EOF
   ```
   {% endraw %}

6. Allow usage of Container Registry without TLS for docker:

   Using menu Docker Desktop -> Settings -> Docker Engine add following configuration key:

   ```json
   {
      "insecure-registries": ["registry.example.com:80"]
   }
   ```

   Restart Docker Desktop using right button menu of the tray Docker Desktop icon.

   Then start minikube again:

   {% raw %}
   ```shell
   minikube start --vm=true --insecure-registry registry.example.com:80
   ```
   {% endraw %}

7. Allow usage of Container Registry without TLS for werf:

   Set `WERF_INSECURE_REGISTRY=1` environment variable in the terminal where werf would run. For bash:

   ```
   export WERF_INSECURE_REGISTRY=1
   ```

   To set this option automatically in new bash-sessions, add it to the `.bashrc`:

   ```shell
   echo export WERF_INSECURE_REGISTRY=1 | tee -a ~/.bashrc
   ```

8. We are going to use `vote.quickstart-application.example.com` and `result.quickstart-application.example.com` domains to access application and `registry.example.com` domain to access Container Registry.

   Let's update hosts file. Run the following command in the terminal:
   
   ```shell
   echo "$(minikube ip) vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com" | sudo tee -a /etc/hosts
   ```

9. Let's also add `registry.example.com` domain to the minikube node:

   ```shell
   minikube ssh -- "echo $(minikube ip) registry.example.com | sudo tee -a /etc/hosts"
   ```
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Linux — minikube</a>
<div class="details__content" markdown="1">

1. Install [minikube](https://github.com/kubernetes/minikube#installation) according to the [instructions](https://minikube.sigs.k8s.io/docs/start/) (completing its first section called "Installation" is sufficient).

2. Start minikube:

   {% raw %}
   ```shell
   minikube start --driver=docker --insecure-registry registry.example.com:80
   ```
   {% endraw %}
    
   **IMPORTANT** Param `--insecure-registry` allows usage of Container Registry without TLS. TLS in our case dropped for simplicity.

3. If you haven't installed `kubectl` yet, you can create an alias to use `kubectl` supplied with minikube:

   ```
   alias kubectl="minikube kubectl --"
   echo 'alias kubectl="minikube kubectl --"' >> ~/.bash_aliases
   ```

4. Install NGINX Ingress Controller:

   {% raw %}
   ```shell
   minikube addons enable ingress
   ```
   {% endraw %}

5. Install Container Registry to store images:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}
    
   Create Ingress to access Container Registry:
 
   {% raw %}
   ```shell
   kubectl apply -f - << EOF
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: registry
     namespace: kube-system
     annotations:
       nginx.ingress.kubernetes.io/proxy-body-size: "0"
   spec:
     rules:
     - host: registry.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: registry
               port:
                 number: 80
   EOF
   ```
   {% endraw %}

6. Allow usage of Container Registry without TLS for Docker:

   Add new key to the file `/etc/docker/daemon.json` (default location):

   ```json
   {
      "insecure-registries": ["registry.example.com:80"]
   }
   ```

   If there is no such file in the directory, you need to create it and insert the above lines into it. Note that you need the superuser (root) privileges to access and modify files in the `/etc` directory.

   Restart Docker:
   
   ```shell
   sudo systemctl restart docker
   ```

   Then start minikube again:

   {% raw %}
   ```shell
   minikube start --driver=docker --insecure-registry registry.example.com:80
   ```
   {% endraw %}

7. Allow usage of Container Registry without TLS for werf:

   Set `WERF_INSECURE_REGISTRY=1` environment variable in the terminal where werf would run. For bash:

   ```
   export WERF_INSECURE_REGISTRY=1
   ```

   To set this option automatically in new bash-sessions, add it to the `.bashrc`:

   ```shell
   echo export WERF_INSECURE_REGISTRY=1 | tee -a ~/.bashrc
   ```

8. We are going to use `vote.quickstart-application.example.com` and `result.quickstart-application.example.com` domains to access application and `registry.example.com` domain to access Container Registry.

   Let's update hosts file. Make sure minikube is up and running:

   ```shell
   echo "$(minikube ip)"
   ```

   If the result shows the IP address of the cluster, your cluster is up.
   
   Run the following command in the terminal:
   
   ```shell
   echo "$(minikube ip) vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com" | sudo tee -a /etc/hosts
   ```

9. Let's also add `registry.example.com` domain to the minikube node:

   ```shell
   minikube ssh -- "echo $(minikube ip) registry.example.com | sudo tee -a /etc/hosts"
   ```
</div>
</div>

## Deploy an example application

1. Clone the example application's repository:

   {% raw %}
   ```shell
   git clone https://github.com/werf/quickstart-application
   cd quickstart-application
   ```
   {% endraw %}

2. Run the converge command using your container registry for storing images (`localhost:5000/quickstart-application` repository in the case of a local container registry).

   {% raw %}
   ```shell
   werf converge --repo registry.example.com:80/quickstart-application
   ```
   {% endraw %}

_NOTE: `werf` uses the same settings to connect to the Kubernetes cluster as the `kubectl` tool does: the `~/.kube/config` file and the `KUBECONFIG` environment variable. werf also supports `--kube-config` and `--kube-config-base64` parameters for specifying custom kubeconfig files._

## Check the result

When the converge command is successfully completed, it is safe to assume that our application is up and running.

Our application is a basic voting system. Let’s check it!

1. Go to the following URL to vote: [vote.quickstart-application.example.com](http://vote.quickstart-application.example.com)

2. Go to the following URL to check the result of voting: [result.quickstart-application.example.com](http://result.quickstart-application.example.com)

## How it works

To deploy an application using werf, we should define the desired state in the Git (as set out in the [How it works](/how_it_works.html)).

1. We have the following Dockerfiles in our repository:

   {% raw %}
   ```
   vote/Dockerfile
   result/Dockerfile
   worker/Dockerfile
   ```
   {% endraw %}

2. The `werf.yaml` file references those Dockerfiles:

   {% raw %}
   ```
   configVersion: 1
   project: quickstart-application
   ---
   image: vote
   dockerfile: Dockerfile
   context: vote
   ---
   image: result
   dockerfile: Dockerfile
   context: result
   ---
   image: worker
   dockerfile: Dockerfile
   context: worker
   ```
   {% endraw %}

3. Kubernetes templates for `vote`, `db`, `redis`, `result` and `worker` components of the application are described in the files of a `.helm/templates/` directory. Components interact with each other as shown in the diagram:

  ![Component interaction diagram]({{ "images/quickstart-architecture.svg" | true_relative_url }})

   - A front-end web app in Python or ASP.NET Core lets you vote for one of the two options;
   - A Redis or NATS queue collects new votes;
   - A .NET Core, Java or .NET Core 2.1 worker consumes votes and stores them in…
   - A Postgres or TiDB database backed by a Docker volume;
   - A Node.js or ASP.NET Core SignalR web-app shows the results of the voting in real-time.

## What's next?

Check the ["Using werf with CI/CD systems" article](using_with_ci_cd_systems.html) or refer to the [guides](/guides.html).
