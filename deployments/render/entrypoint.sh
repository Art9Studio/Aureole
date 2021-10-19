#!/bin/bash

set -e

# cd to workdir
cd app

# migrate db schema
psql $DATABASE_URL -f render/schema.sql

# if apple key not empty, save it to file
if [ ! -z "$APPLE_PRIVATE_KEY" ];
then
  begin_part=$(echo $APPLE_PRIVATE_KEY | grep -Eo -- "-+[A-Za-z ]+-+" | awk 'NR==1{print}')
  end_part=$(echo $APPLE_PRIVATE_KEY | grep -Eo -- "-+[A-Za-z ]+-+" | awk 'NR==2{print}')

  raw_main_part=$(printf '%s\n' "${APPLE_PRIVATE_KEY//$begin_part /}")
  raw_main_part=$(printf '%s\n' "${raw_main_part// $end_part/}")

  main_part="${raw_main_part// /\\n}"
  full_key="$begin_part\n$main_part\n$end_part"

  echo -e $full_key > "render/res/apple_key.p8"
fi

# launch aureole
./aureole