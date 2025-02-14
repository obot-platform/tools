# Wordpress Tools

## Development with Wordpress API
### (optional) Run Wordpress locally with docker:
create a yaml file `wordpress.yaml`:

```yaml
services:

  wordpress:
    image: wordpress
    restart: always
    ports:
      - 8070:80
    environment:
      WORDPRESS_DB_HOST: db
      WORDPRESS_DB_USER: dbuser
      WORDPRESS_DB_PASSWORD: dbpassword
      WORDPRESS_DB_NAME: exampledb
      WORDPRESS_CONFIG_EXTRA: |
        define('WP_ENVIRONMENT_TYPE', 'local');
    volumes:
      - wordpress:/var/www/html

  db:
    image: mysql:8.0
    restart: always
    environment:
      MYSQL_DATABASE: exampledb
      MYSQL_USER: dbuser
      MYSQL_PASSWORD: dbpassword
      MYSQL_RANDOM_ROOT_PASSWORD: '1'
    volumes:
      - db:/var/lib/mysql

volumes:
  wordpress:
  db:

```
and run `docker compose -f wordpress.yaml up`.

Navigate to localhost:8070. The first time you run the container, it will prompt you to register the site name and create a new user.

### Obot Basic Auth Configuration

- username: the username you created first time you nagivate to the site
- password: the password of the user you created first time you nagivate to the site
- url: the url of your wordpress site. For local, it's `http://localhost:8070`. For hosted, it's the url of your wordpress site, such as `https://example.wpenginepowered.com`.

### CRITICAL: Configure Permalinks in Settings
1. Without configuring permalinks, the Wordpress API will not work, because the `/wp-json` endpoint will not be available.
To do this, go to your wordpress site dashboard, on the left sidebar, select `Settings` -> `Permalinks`, then select any non-plain Permalink structure.
2. You must create an application password to be able to create posts. 
To do this, go to your wordpress site dashboard, on the left sidebar, select `Users`, edit your user profile and scroll down to the `Application Passwords` section, then select `Add New`.

