# Pokédex web

Ce projet est la reprise et l’évolution web d’un ancien projet Pokédex. Il permet de consulter les données de l’API publique [PokéAPI](https://pokeapi.co/). L’interface est en HTML et CSS ; le serveur Go effectue les recherches, sans JavaScript ni bibliothèque externe.

Prérequis : Go 1.27.1 ou ultérieur et une connexion Internet.

## Utiliser le site

Depuis le dossier du projet, lancez le serveur dans CMD ou PowerShell :

```sh
go run .
```

Ouvrez ensuite [http://localhost:8080](http://localhost:8080) dans un navigateur. Le terminal sert uniquement à démarrer le serveur ; les fonctionnalités s’utilisent dans le navigateur. Arrêtez le serveur avec `Ctrl+C`.

Le site permet de :

- rechercher un Pokémon par nom ou identifiant ;
- afficher ses caractéristiques, types, statistiques, sprite et faiblesses de type ;
- ouvrir une page de préparation de combat avec deux champs de recherche et un bouton VS.

Le champ de recherche propose les noms disponibles avec l’autocomplétion native du navigateur, sans JavaScript.
Le bouton VS compare les six statistiques de base, ainsi que la taille et le poids. Chaque valeur supérieure rapporte un point ; une égalité ne rapporte rien. Le Pokémon dont le multiplicateur de dégâts contre l’autre est le plus élevé rapporte deux points par statistique gagnée. Si les multiplicateurs sont égaux et supérieurs à 1, les deux Pokémon ont ce bonus. Le score le plus élevé détermine le gagnant.

Le tirage aléatoire et les anciennes commandes d’affichage dans le terminal ont été retirés. Une connexion Internet est nécessaire pour accéder à PokéAPI.

## Arborescence

```text
.
├──  historique-conversation.pdf
├── go.mod
├── main.go
├── internal/
│   ├── api/                 # Appels à PokéAPI et tests
│   ├── pokemon/             # Modèles, règles de combat et tests
│   └── web/
│       ├── server.go        # Routes et rendu des pages
│       ├── server_test.go
│       ├── static/          # Feuilles de style CSS
│       │   ├── combat.css
│       │   └── style.css
│       └── templates/       # Modèles HTML Go
│           ├── combat.html
│           ├── index.html
│           └── pokemon-info.html
```

Les fichiers HTML et CSS sont intégrés au binaire Go, ce qui permet de lancer le serveur sans dépendre de leur emplacement sur le disque.

Le PDF `docs/historique-conversation.pdf` contient la transcription des échanges enregistrés pour ce projet.

## Vérifier le projet

```sh
go test ./...
```
