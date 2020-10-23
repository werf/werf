### Docker

[Docker CE installation guide](https://docs.docker.com/install/).

Чтобы получить доступ к использованию docker-server для текущего пользователя системы, необходимо добавить его в группу `docker`:

```shell
sudo groupadd docker
sudo usermod -aG docker $USER
```

### Утилита Git

[Git installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

 - Минимальная требуемая версия: 1.9.0.
 - Минимальная версия для использования функционала [Git Submodules](https://git-scm.com/docs/gitsubmodules): 2.14.0
