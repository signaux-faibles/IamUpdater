#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

docker stop keycloak || echo "container keycloak is not running"
docker rm keycloak || echo "container keycloak doesn't exist"
docker run -p 8080:8080 --name keycloak ghcr.io/signaux-faibles/conteneurs/keycloak:v1.0.0 > /dev/null &
sleep 20
docker exec keycloak /opt/jboss/keycloak/bin/add-user-keycloak.sh -u kcadmin -p kcpwd
docker restart keycloak
sleep 20
echo "Keycloak is ready with user 'kcadmin' provided"