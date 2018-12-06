# migrator docker

To run migrator docker you need to:

* pull `lukasz/migrator` from docker cloud
* mount a volume with migrations under `/data`
* (optional) specify location of migrator configuration file via environmental variable `MIGRATOR_YAML`, defaults to `/data/migrator.yaml`

To run migrator as a service:

```bash
docker run -p 8080:8080 -v /Users/lukasz/migrator-test:/data -e MIGRATOR_YAML=/data/m.yaml -d --link migrator-postgres lukasz/migrator
Starting migrator using config file: /data/m.yaml
2016/08/04 06:24:58 Read config file ==> OK
2016/08/04 06:24:58 Migrator web server starting on port 8080...
```

To run migrator in interactive terminal mode:

```bash
docker run -it -v /Users/lukasz/migrator-test:/data --entrypoint sh --link migrator-postgres lukasz/migrator
```

# History and releases

Here is a short history of migrator docker images:

1. initial release in 2016 - migrator on debian:jessie - 603MB
2. v1.0 - migrator v1.0 on golang:1.11.2-alpine3.8 - 346MB
3. v1.0-mini - migrator v1.0 multi-stage build with final image on alpine:3.8 - 13.4MB
4. v2.0 - migrator v2.0 - 14.8MB

Starting with v2.0 all migrator images by default use multi-stage builds. For migrator v1.0 you have to explicitly use `v1.0-mini` tag in order to enjoy an ultra lightweight migrator image. Still, I recommend using latest and greatest.

Finally, starting with v2.0 migrator-docker project was merged into migrator main project. New version of docker image is built automatically every time a new release is created.

To view all available docker containers see [lukasz/migrator/tags](https://cloud.docker.com/repository/docker/lukasz/migrator/tags).

# License

Copyright 2016-2018 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
