# `Standuply Test`
## `Install`
1. Copy settings.json: 
    ```shell script
    $ cp settings.json.example settings.json
    $ vim settings.json
    ```
    Edit database connection parameters and key API Slack.
1. Build app:
    ```shell script
    $ go build
    ```
1. Synchronize database model:
    ```shell script
    $ ./standuplytest -sync-db
    ```
1. Run app:
    ```shell script
    $ ./standuplytest
    ```