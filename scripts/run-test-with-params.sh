#!/bin/bash

for i in "GOAT_EMAIL_HOST" \
"GOAT_EMAIL_PORT" \
"GOAT_EMAIL_USER" \
"GOAT_EMAIL_PASSWORD" \
"GOAT_EMAIL_FROM" \
"GOAT_EMAIL_TO"; do
    read -p "Enter $i: " $i
    eval "export $i=\${$i}"
done

make test
