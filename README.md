# KeycloakUpdater pour signaux-faibles
Outil de maintenance la base utilisateurs Keycloak à jour à partir d'un fichier excel
Cet outil est spécifique pour le projet Signaux-Faibles.

## Installation
```
git clone https://github.com/signaux-faibles/keycloakUpdater
go build
```

## Préparation de Keycloak
Si vous ne souhaitez pas utiliser le realm master, créez en un autre et spécifiez son nom dans `config.toml`

Créez un client avec le nom de votre choix, en prenant soin de spécifier ce nom dans `config.toml`

## Configuration
La configuration de keycloakUpdater se fait à l'aide du fichier `config.toml` 
et des éventuels fichiers `toml` ajoutés dans le répertoire `./config.d`.  
__ATTENTION :__ 
- les paramètres identiques dans plusieurs fichiers seront écrasés par ceux du dernier fichier pris en compte
dans `config.d`. 
- Un tableau ne peut pas être partagé entre plusieurs fichiers.
- `Viper` gère les paramètres en `lower case` sauf dans les maps, donc les paramètres sont ceux de l'API `Keycloak` 
  mais il faut remplacer les majuscules par un `_` avec la lettre en minuscule.
  Ex : `ImplicitFlowEnabled` devient `implicit_flow_enabled`

Le fichier [`config.toml.example`](config.toml.example) est un exemple de configuration. 
Le compte utilisateur `keycloak` stipulé doit avoir l'intégralité des droits d'administration sur le serveur keycloak.

### Configuration d'un client'
La configuration des clients se fait avec la balise 
(soit dans le fichier `config.toml` soit dans un fichier dédié dans le répertoire `config.d`)
```toml
[[client.nomduclient]]
```
Celà suffit à créer un client.
Pour la configuration complémentaire, il faut ajouter sous cette balise les paramètres de [`ClientRepresentation`
dans l'API Keycloak](https://www.keycloak.org/docs-api/17.0/rest-api/index.html#_clientrepresentation).
__ATTENTION :__ tout n'est pas encore développé.

### Configuration des utilisateurs
Renseignez la base utilisateur dans le fichier excel fourni (userBase.xlsx), le chemin peut être ajusté dans `config.toml`.

## les TestsPour tester via `Docker`
### Lancer Keycloak en `localhost` via Docker
```bash
# lancer le container
docker run -p 8080:8080 --name keycloak ghcr.io/signaux-faibles/conteneurs/keycloak:v1.0.0
# créer l'utilisateur 
# (chez moi, les arguments pour créer un utilisateur au lancement du container ne fonctionnent pas)
# ce user doit être déclaré dans le fichier `excel` des users
docker exec keycloak  /opt/jboss/keycloak/bin/add-user-keycloak.sh -u kcadmin -p kcpwd
# redémarrer
docker restart keycloak
```

### Lancer les tests `go`
- Lancer les tests dans tous les packages
  ```bash
  go test ./...
  ```
- Lancer le test d'intégration
  ```bash
  go test -tags=integration
  ```
## Erreurs
- `panic: 401 Unauthorized: invalid_grant: Invalid user credentials` 
  -> il faut s'assurer que le user est bien créé dans `Keycloak`, qu'il a les droits nécessaires
  et que ses credentials sont bien configurés dans le fichier de config (voir `## Pour tester via Docker`)
## Format Excel
Le niveau peut prendre 0, A ou B:
- 0: utilisateur administratif (compte de service, ou administrateur)
- A: utilisateur de l'application niveau A
- B: utilisateur de l'application niveau B

La zone géographique peut être le numéro d'un département, ou d'une région renseignée dans l'onglet `zones` du fichier excel.

## Licence
Copyright © 09/25/2020, Christophe Ninucci, Rapha

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

The Software is provided “as is”, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders X be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the Software.
Except as contained in this notice, the name of Christophe Ninucci shall not be used in advertising or otherwise to promote the sale, use or other dealings in this Software without prior written authorization from Christophe Ninucci.
