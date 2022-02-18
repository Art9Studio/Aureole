#!/bin/bash

set -e

cd app

if [ ! -z "$APPLE_PRIVATE_KEY" ];
then
      printf '%s\n' "$APPLE_PRIVATE_KEY" > keys/apple_private_key.p8
fi

if [[ -z "$APP_HOST" && ! -z "$HEROKU_APP_NAME" ]]
then
      export APP_HOST="https://$HEROKU_APP_NAME.herokuapp.com/"
fi

./aureole