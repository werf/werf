
Однако базовый образ `maven:3-jdk-8`, который мы использовали для сборки, достаточно тяжелый, нет смысла запускать код со всеми сборочными зависимостями в kubernetes. 
Будем запускать используя достаточно легкий `openjdk:8-jdk-alpine`. Но нам все еще нужно собирать jar в образе с maven. Для реализации этого решения воспользуемся [артефактом](https://werf.io/documentation/configuration/stapel_artifact.html). По сути это то же самое что и `image` в директивах `werf.yaml`, только временный. Он не пушится в registry.
Переименуем `image` в `artifact` и назовем его `build`. Результатом его работы является собранный jar - именно его мы и импортируем в `image` с `alpine-openjdk` который опишем в этом же `werf.yaml` после "---", которые разделяют `image`. Назовем его spring и уже его пушнем в registry для последующего запуска в кластере.

```yaml
---
image: basicapp
from: openjdk:8-jdk-alpine
import:
- artifact: build
  add: /app/target/*.jar
  to: /app/demo.jar
  after: setup
```

[werf.yaml](gitlab-java-springboot-files/01-demo-optimization/werf.yaml:32-39)

Для импорта между `image` и `artifact` служит директива `import`. Из `/app/target` в сборочном артефакте импортируем собранный jar-файл в папку /app в image spring. Единственное что следует еще поправить - это версию собираемого jar в [pom.xml](01-demo-optimization/pom.xml:14). Пропишем её 1.0, чтобы имя итогового jar-файла получось предсказуемым - demo-1.0.jar. 

