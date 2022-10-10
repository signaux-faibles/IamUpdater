[![CI](https://github.com/signaux-faibles/keycloakUpdater/actions/workflows/pipeline.yml/badge.svg)](https://github.com/signaux-faibles/keycloakUpdater/actions/workflows/pipeline.yml)

# KeycloakUpdater pour signaux-faibles
Cet outil permet :
- de maintenir à jour à la base des utilisateurs Keycloak ainsi que leurs droits à partir d'un fichier excel.
- de configurer le realm et les clients Keycloak nécessaires au bon fonctionnement de la plateforme
- de créer et habiliter les utilisateurs dans la base Wekan sur les tableaux

- Cet outil est spécifique pour le projet Signaux-Faibles.

## Installation
```
git clone https://github.com/signaux-faibles/keycloakUpdater
go build
```

## Configuration
La configuration de keycloakUpdater se fait à l'aide du fichier `config.toml`.

Voir [l'exemple](/test/sample) pour plus de précisions. 
On voit qu'il y a 3 sections à remplir
- [keycloak] contenant les informations d'accès à keycloak
- [logger] contenant la configuration du fichier de log
- [stock] contenant le chemin vers le répertoire où seront posés les fichiers de configuration des clients et du realm ainsi que le fichier stock des users
- [wekan] contenant les informations de connexion à Wekan



### Configuration d'un client/ du realm
Il faut poser un fichier `toml` dans le répertoire précisé dans la balise `clientsAndRealmFolder` de la section `stock`
du fichier principal de configuration. 
Ce fichier doit contenir les paramètres [documentés](https://www.keycloak.org/docs-api/18.0/rest-api/) dans l'API d'administration de Keycloak.
Le realm correspond à [`RealmRepresentation`](https://www.keycloak.org/docs-api/18.0/rest-api/#_realmrepresentation)
Un client correspond à [`ClientRepresentation`](https://www.keycloak.org/docs-api/18.0/rest-api/#_clientrepresentation)

### Configuration des utilisateurs
Renseignez la base utilisateur dans le fichier excel fourni (userBase.xlsx), le chemin peut être ajusté dans `config.toml`.


### Lancer les tests `go`
- Lancer les tests dans tous les packages
  ```bash
  go test ./...
  ```
- Lancer le test d'intégration
  ```bash
  go test -tags=integration
  ```
  __ATTENTION :__ 
- si les tests d'intégration sont lancés via un IDE, ils seront lancés parallèlement,
  ce qui causera des soucis. Pour éviter ça il faut rajouter le paramètre `-p 1` pour que ce soit lancé à la suite.
- 2 containers keycloak ne peuvent pas tourner en même temps à cause d'`Infinispan`, voir l'erreur ci dessous
  > Could not connect to keycloak: reached retry deadline
   

### Tester localement
```bash
# 1. Lancer le conteneur keycloak
docker run -p 8080:8080 --name keycloak --env KEYCLOAK_USER=kcadmin --env KEYCLOAK_PASSWORD=kcpwd ghcr.io/signaux-faibles/conteneurs/keycloak:v1.0.0
# 2. Créer le conteneur (avec le tag `ku`)
docker build  --tag ku .
# 3. Lancer le conteneur avec la configuration qui va bien
docker run --rm --name ku --volume /path/to/keycloakUpdater/test/sample:/workspace --link keycloak:keycloak ku
```
Il faut monter un volume avec les fichiers de configuration dans le répertoire `workspace` du conteneur.


## Erreurs
- `panic: 401 Unauthorized: invalid_grant: Invalid user credentials` 
  -> il faut s'assurer que l'utilisateur est bien créé dans `Keycloak`, qu'il a les droits nécessaires
  et que ses credentials sont bien configurés dans le fichier de config (voir `## Pour tester via Docker`)
- `PANIC  [2022-05-11 11:06:18] Could not connect to keycloak: reached retry deadline`
  Il peut s'agir d'un problème de version avec un autre conteneur Keycloak. Il faut stopper l'autre conteneur ou aligner les versions.
## Format Excel
Le niveau peut prendre 0, A ou B :
- 0 : utilisateur administratif (compte de service ou administrateur)
- A : utilisateur de l'application niveau A
- B : utilisateur de l'application niveau B

La zone géographique peut être le numéro d'un département, ou d'une région renseignée dans l'onglet `zones` du fichier excel.

## Licence
Copyright © 09/25/2020, Christophe Ninucci, Raphaël Squelbut

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

The Software is provided “as is”, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders X be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the Software.
Except as contained in this notice, the name of Christophe Ninucci shall not be used in advertising or otherwise to promote the sale, use or other dealings in this Software without prior written authorization from Christophe Ninucci.
