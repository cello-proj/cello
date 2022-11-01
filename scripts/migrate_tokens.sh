#!/bin/bash

set -eu

extract_status_code () {
  echo $1 | awk '{print $NF}'
}

extract_body () {
  echo ${1::-3}
}

: "${PGHOST?PGHOST must be set}"
: "${PGPORT?PGPORT must be set}"
: "${PGDATABASE?PGDATABASE must be set}"
: "${PGUSER?PGUSER must be set}"
: "${VAULT_HOST?VAULT_HOST must be set}"
: "${VAULT_TOKEN?VAULT_TOKEN must be set}"

vault_token=$VAULT_TOKEN
vault_host=$VAULT_HOST

# get list of projects/approles
approle_response=$(curl -s \
	--header "X-Vault-Token: $vault_token" \
	--request LIST \
	"$vault_host/v1/auth/approle/role" \
	| jq -r '.data.keys')

for approle in $(echo "$approle_response" | jq -r '.[]'); do
  # get list of accessors for each approle/project
  accessors_resp=$(curl -s -w "%{http_code}" \
    --header "X-Vault-Token: $vault_token" \
    --request LIST \
    "$vault_host/v1/auth/approle/role/$approle/secret-id" \
    | jq .)

  accessors_status_code=$(extract_status_code "$accessors_resp")
  if [ "$accessors_status_code" != "200" ]; then
	  echo "ERROR: list accessors failed: $approle status_code: $accessors_status_code"
	  continue
  fi

  accessors_body=$(extract_body "$accessors_resp")
  accessors=($(echo $accessors_body | jq -r '.data.keys' | tr -d '[]," '))

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
	project=${approle#"argo-cloudops-projects-"}

	exists=$(psql -h $PGHOST -p $PGPORT -d $PGDATABASE -U $PGUSER -w -t -c "SELECT EXISTS(SELECT project from projects where project='$project')")
	if [ $exists == "f" ]; then
		# due to a foreign key constraint, inserting into tokens table will fail if the project does not exist in the projects table 
		echo "ERROR: Project does not exist in table: $project"
		continue
	fi

	# get created time and ttl of secret or token
	lookup_resp=$(curl -s -w "%{http_code}" \
		--header "X-Vault-Token: $vault_token" \
		--data "{\"secret_id_accessor\":\"${accessors[0]}\"}" \
		--request POST \
		"$vault_host/v1/auth/approle/role/$approle/secret-id-accessor/lookup" \
		| jq .)

        lookup_status_code=$(extract_status_code "$lookup_resp")
	if [ "$lookup_status_code" != "200" ]; then
		echo "ERROR: lookup accessor failed: $approle status_code: $lookup_status_code"
		continue
	fi

	accessor_data=$(extract_body "$lookup_resp")

	creation_time=$(echo $accessor_data | jq -r '.data.creation_time')
	expiration_time=$(echo $accessor_data | jq -r '.data.expiration_time')

	set +e
	echo "Inserting accessor ID into Tokens table for project: $project"
	psql -q -h $PGHOST -p $PGPORT -d $PGDATABASE -U $PGUSER -w -c "INSERT INTO tokens(token_id, created_at, project) VALUES ('${accessors[0]}', TO_TIMESTAMP('$creation_time', 'YYYY-MM-DD HH24:MI:SSSSZ'), '$project') ON CONFLICT (token_id) DO UPDATE SET created_at='$creation_time', expires_at='$expiration_time';"
	set -e
done

echo "Script complete"
