#! /usr/bin/env bash

# abort on nonzero exitstatus
set -o errexit
# abort on unbound variable
set -o nounset
# don't hide errors within pipes
set -o pipefail

echo "building..."
go build
cp keycloakUpdater test/localhost

echo "starting keycloak..."
docker stop keycloak || echo "keycloak non started"
docker rm keycloak || echo "keycloak don't exist"
docker run --detach --publish 8080:8080 --name keycloak --env KEYCLOAK_USER=kcadmin --env KEYCLOAK_PASSWORD=kcpwd ghcr.io/signaux-faibles/conteneurs/keycloak:v1.0.0
#docker run --detach --publish 8080:8080 --name keycloak --env KEYCLOAK_USER=kcadmin --env KEYCLOAK_PASSWORD=kcpwd docker pull jboss/keycloak:16.1.1
echo "waiting 15s while keycloak is starting..."
sleep 15

echo "starting keycloakUpdater..."
cd test/localhost/
./keycloakUpdater
cd -