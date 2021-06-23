# Common Docker build images

## Authentication and project setup

```
gcloud config configurations activate vb-staging
gcloud auth configure-docker
```

(Or use `vb-production` for production...)

## Base builder

Rebuild when:
  * Builder OS package requirements change (i.e. when
    `build/Dockerfile.base-builder` file has been changed)

```
cd .../backend
docker build -t base-builder -f build/Dockerfile.base-builder .
docker tag base-builder eu.gcr.io/veganbase/base-builder
docker push eu.gcr.io/veganbase/base-builder
```

## Module builder

Rebuild when:
  * `go.mod` or `go.sum` changes; OR
  * Base builder is rebuilt

```
cd .../backend
docker build -t module-builder -f build/Dockerfile.module-builder .
docker tag module-builder eu.gcr.io/veganbase/module-builder
docker push eu.gcr.io/veganbase/module-builder
```

## Service builder

Rebuild when:
  * Server source code changes; OR
  * Code generation sources change; OR
  * Module builder is rebuilt

```
cd .../backend
docker build -t service-builder -f build/Dockerfile.service-builder .
docker tag service-builder eu.gcr.io/veganbase/service-builder
docker push eu.gcr.io/veganbase/service-builder
```
