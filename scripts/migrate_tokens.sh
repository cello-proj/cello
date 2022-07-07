#!/bin/bash

set -eu
: "${PGHOST?PGHOST must be set}"
: "${PGPORT?PGPORT must be set}"
: "${PGDATABASE?PGDATABASE must be set}"
: "${PGUSER?PGUSER must be set}"
: "${VAULT_HOST?VAULT_HOST must be set}"
: "${VAULT_TOKEN?VAULT_TOKEN must be set}"

vault_token=$VAULT_TOKEN
vault_host=$VAULT_HOST

# get list of projects/approles
approle_response=$(curl -s\
	--header "X-Vault-Token: $vault_token" \
	--request LIST \
	"$vault_host/v1/auth/approle/role" \
	| jq -r '.data.keys')

for approle in $(echo "$approle_response" | jq -r '.[]'); do
  # get list of accessors for each approle/project
  accessors=($(curl -s\
    --header "X-Vault-Token: $vault_token" \
    --request LIST \
    "$vault_host/v1/auth/approle/role/$approle/secret-id" \
    | jq -r '.data.keys' | tr -d '[]," '))

  # if multiple accessors exist, print approle for manual triage.
  if [ "${#accessors[@]}" -gt 1 ]; then
		echo "ERROR: Multiple accessors exist for AppRole: $approle"
		continue
  fi

  if [ "${#accessors[@]}" -eq 0 ]; then
		echo "ERROR: Accessors do not exist for AppRole: $approle"
		continue
  fi

	# if any modifications are needed to the project name for the DB, it should happen here
	project=$approle

	exists=$(psql -h $PGHOST -p $PGPORT -d $PGDATABASE -U $PGUSER -w -t -c "SELECT EXISTS(SELECT project from projects where project='$project')")
	if [ $exists == "f" ]; then
		# due to a foreign key constraint, inserting into tokens table will fail if the project does not exist in the projects table 
		echo "ERROR: Project does not exist in table: $project"
		continue
	fi

	set +e
	echo "Inserting accessor ID into Tokens table for AppRole: $approle"
	psql -h $PGHOST -p $PGPORT -d $PGDATABASE -U $PGUSER -w <<'SQL' | ...
	  INSERT INTO tokens(token_id, created_at, project)
		VALUES ('${accessors[0]}', current_timestamp, '$approle')
SQL
	set -e
done

echo "Script complete"
