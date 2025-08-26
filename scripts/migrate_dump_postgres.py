# /// script
# requires-python = ">=3.12"
# dependencies = []
# ///
"""
PostgreSQL to DynamoDB Migration - Data Dump Script

This script dumps project and token data from PostgreSQL to JSON files
for later import into DynamoDB.

Usage:
    uv run scripts/migrate_dump_postgres.py <db_host[:port]> <db_user> <db_name> <output_file>
    
    Examples:
    - db_host: localhost (uses default port 5432)
    - db_host: localhost:5433 (uses custom port 5433)
    - db_host: db.example.com:5432 (uses custom port 5432)
    
    Password can be provided via PGPASSWORD environment variable:
    PGPASSWORD=mypassword uv run scripts/migrate_dump_postgres.py <db_host[:port]> <db_user> <db_name> <output_file>
"""

import sys
import subprocess
import json
import argparse
import os
from pathlib import Path


def parse_host_port(host_arg):
    """Parse host:port format and return (host, port)."""
    if ':' in host_arg:
        host, port_str = host_arg.rsplit(':', 1)
        try:
            port = int(port_str)
            return host, port
        except ValueError:
            print(f"✗ Invalid port number: {port_str}")
            sys.exit(1)
    else:
        return host_arg, None


def run_psql_query(db_host, db_user, db_name, db_port, password, query):
    """Execute a PostgreSQL query and return results as a list of dictionaries."""
    cmd = [
        "psql",
        "-h", db_host,
        "-U", db_user,
        "-d", db_name,
    ]
    if db_port:
        cmd.extend(["-p", str(db_port)])
    cmd.extend([
        "-c", query,
        "-A",  # No alignment
        "-t",  # Tuples only
        "-F", ",",  # CSV format
    ])

    try:
        # If password is provided, set it in the environment for this subprocess only
        env = os.environ.copy()
        if password:
            env["PGPASSWORD"] = password
            
        result = subprocess.run(
            cmd, capture_output=True, text=True, check=True, env=env
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"✗ Error executing query: {e}")
        print(f"stderr: {e.stderr}")
        sys.exit(1)


def parse_csv_output(csv_output):
    """Parse CSV output from psql into a list of dictionaries."""
    lines = csv_output.strip().split('\n')
    if not lines or lines[0] == '':
        return []
    
    # PostgreSQL with -A -t -F ',' doesn't return headers, so we need to define them
    # based on the query structure
    data = []
    
    for line in lines:
        if line.strip():
            values = line.split(',')
            if len(values) >= 2:  # Ensure we have at least 2 columns
                # For projects: project, repository
                if len(values) == 2:
                    row = {
                        'project': values[0],
                        'repository': values[1]
                    }
                    data.append(row)
                # For tokens: token_id, created_at, project, expires_at
                elif len(values) == 4:
                    row = {
                        'token_id': values[0],
                        'created_at': values[1],
                        'project': values[2],
                        'expires_at': values[3]
                    }
                    data.append(row)
    
    return data


def main():
    parser = argparse.ArgumentParser(
        description="Dump PostgreSQL data to JSON for DynamoDB migration",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Basic usage with default port (5432) - password will be prompted
  uv run scripts/migrate_dump_postgres.py localhost cello_user cello_db migration_data.json
  
  # With custom port using host:port format
  uv run scripts/migrate_dump_postgres.py localhost:5433 cello_user cello_db migration_data.json
  
  # With remote server and custom port
  uv run scripts/migrate_dump_postgres.py db.example.com:5432 cello_user cello_db migration_data.json
  
  # Using PGPASSWORD environment variable for password
  PGPASSWORD=mypassword uv run scripts/migrate_dump_postgres.py localhost cello_user cello_db migration_data.json

Note: If PGPASSWORD environment variable is not set, the script will securely prompt for the password.
        """
    )
    
    parser.add_argument("db_host", help="PostgreSQL server hostname or IP address (optionally with :port)")
    parser.add_argument("db_user", help="PostgreSQL username")
    parser.add_argument("db_name", help="PostgreSQL database name")
    parser.add_argument("output_file", help="Path to the output JSON file")
    
    args = parser.parse_args()
    
    # Parse host and port from the db_host argument
    db_host, db_port = parse_host_port(args.db_host)
    
    # Get password from environment or prompt securely
    password = None
    if "PGPASSWORD" in os.environ:
        password = os.environ["PGPASSWORD"]
    else:
        import getpass
        password = getpass.getpass("Enter PostgreSQL password: ")
        if not password:
            print("✗ Password is required")
            sys.exit(1)

    # Create output directory if it doesn't exist
    output_path = Path(args.output_file)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    print("Fetching projects data...")
    projects_query = """
    SELECT project, repository 
    FROM projects 
    ORDER BY project;
    """
    projects_csv = run_psql_query(db_host, args.db_user, args.db_name, db_port, password, projects_query)
    projects_data = parse_csv_output(projects_csv)

    print("Fetching tokens data...")
    tokens_query = """
    SELECT token_id, created_at, project, expires_at 
    FROM tokens 
    ORDER BY project, token_id;
    """
    tokens_csv = run_psql_query(db_host, args.db_user, args.db_name, db_port, password, tokens_query)
    tokens_data = parse_csv_output(tokens_csv)

    # Build the JSON structure
    print("Building JSON structure...")
    migration_data = {}
    
    # Process projects
    for project_row in projects_data:
        project_name = project_row['project']
        migration_data[project_name] = {
            'repository': project_row['repository'],
            'tokens': []
        }
    
    # Process tokens
    for token_row in tokens_data:
        project_name = token_row['project']
        if project_name in migration_data:
            token_data = {
                'token_id': token_row['token_id'],
                'created_at': token_row['created_at'],
                'expires_at': token_row['expires_at']
            }
            migration_data[project_name]['tokens'].append(token_data)

    # Write JSON file
    print(f"Writing data to {args.output_file}...")
    with open(args.output_file, 'w') as f:
        json.dump(migration_data, f, indent=2)

    print("\n✓ Data dump completed successfully!")
    print(f"File created: {args.output_file}")
    print(f"Projects processed: {len(migration_data)}")
    total_tokens = sum(len(project['tokens']) for project in migration_data.values())
    print(f"Cello tokens processed: {total_tokens}")


if __name__ == "__main__":
    main()