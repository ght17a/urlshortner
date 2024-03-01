# URL Shortner - Valentin Hubert Nicolas

## Introduction
L'URL Shortner est un outil déployé en local qui permet de raccourcir des liens longs et complexes en liens courts et clairs, afin de les partager simplement. L'URL raccourcie peut être cliquée et permettra de se rediriger vers le lien originel. Ces URL sont stockées dans une base de données et chaque ID se compose de 6 caractères alpha-numériques uniques et aléatoires.

## Pré-requis
Pour une bonne configuration du projet, il faudra plusieurs choses:
- Un éditeur de code (nous utiliserons Visual Studio Code)
- Golang (le langage utilisé)
- MySQL Workbench (pour la base de données)
- Packages

Pour installer les packages sur Go, vous devez utiliser la commande : ```go get```
Par exemple, pour installer le package bcrypt, vous devrez lancer la commande : ```go get golang.org/x/crypto/bcrypt```

 ## Lancer le projet
 Vous aurez à créer une base de données sur MySQL Workbench. 
 Nous vous préconisons de créer un serveur avec les ports suivants : ```127.0.0.1``` et ```3306```
 Pour lancer le projet, vous devrez vous déplacer au dossier racine du projet. Une fois arrivé, vous devez changer les identifiants, mot de passe et ports serveurs 
 si vous en avez des différents.
 Ensuite, lancez la commande ```go run .\main.go``` pour lancer le serveur local. Si c'est bon, vous aurez un message ```URL Shortener is running on :3030```.

 Pour visualiser notre application, ouvrez votre navigateur favori (nous utiliserons Chrome) et entrez l'URL suivante : ```localhost:3030```.

 L'application est donc lancée et vous pourrez la tester.

 ## Fonctionnalités de notre application
Créer une URL raccourcie : Accédez à localhost:3030 et entrez l'URL que vous souhaitez raccourcir.
Inscription : Pour créer un compte, cliquez sur le lien "Register" et remplissez les informations nécessaires.
Connexion : Si vous avez déjà un compte, cliquez sur le lien "Login" pour accéder à votre espace utilisateur.
Statistiques : Une fois connecté, vous pouvez voir le nombre total de liens générés et le nombre de clics par lien.

 ## Configuration de la base de données
Avant de lancer l'application, il faut s'assurer d'avoir une bonne configuration dans MySQL Workbench. Nous avons créé une base de données nommée urlshortner et nous avons lancé un script SQL pour créer les tables ```users``` et ```urls```, le voici :

```sql
CREATE TABLE urls (
    id INT AUTO_INCREMENT PRIMARY KEY,
    short_key VARCHAR(6) NOT NULL,
    original_url VARCHAR(2048) NOT NULL,
    shortened_url VARCHAR(2048) NOT NULL,
    get_clicked INT NOT NULL DEFAULT 0
);

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);
