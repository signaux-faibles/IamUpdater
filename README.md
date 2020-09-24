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

## Utilisation
Recopiez le fichier `config.toml.example` en `config.toml` et renseignez les informations demandées. Le compte utilisateur keycloak stipulé dans ce fichier doit avoir l'intégralité des droits d'administration sur le serveur keycloak.

Renseignez la base utilisateur dans le fichier excel fourni (userBase.xlsx), le chemin peut être ajusté dans `config.toml`.

## Format Excel
Le niveau peut prendre 0, A ou B:
- 0: utilisateur administratif (compte de service, ou administrateur)
- A: utilisateur de l'application niveau A
- B: utilisateur de l'application niveau B

La zone géographique peut être le numéro d'un département, ou d'une région renseignée dans l'onglet `zones` du fichier excel.

## Licence
Copyright © 09/25/2020, Christophe Ninucci

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

The Software is provided “as is”, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders X be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the Software.
Except as contained in this notice, the name of Christophe Ninucci shall not be used in advertising or otherwise to promote the sale, use or other dealings in this Software without prior written authorization from Christophe Ninucci.