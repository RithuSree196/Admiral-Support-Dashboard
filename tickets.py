import os
from azure.cosmos import CosmosClient, exceptions
from datetime import datetime, timezone, timedelta
import json

# Fetch sensitive information from environment variables
url = os.getenv("COSMOS_URL")
key = os.getenv("COSMOS_KEY")
database_name = os.getenv("DB_NAME")
container_name = os.getenv("CONTAINER_NAME")

# Ensure the environment variables are set
if not url or not key:
    raise ValueError("COSMOS_URL or COSMOS_KEY environment variables are not set!")

# Initialize Cosmos Client and container
client = CosmosClient(url, credential=key)
database = client.get_database_client(database_name)
container = database.get_container_client(container_name)

# Query to fetch ticket details 
details_query = f"""
SELECT * FROM c
"""

try:
    # Execute the query to fetch ticket details
    ticket_details = list(container.query_items(
        query=details_query,
        enable_cross_partition_query=True
    ))

    # Process and save the results
    if ticket_details:
        # Save the fetched tickets to a JSON file
        output_file = "source_data.json"
        with open(output_file, "w", encoding="utf-8") as json_file:
            json.dump(ticket_details, json_file, indent=4, default=str)

        print(f"Ticket details saved to {output_file}")
    else:
        print("No tickets found.")

except exceptions.CosmosHttpResponseError as e:
    print(f"An error occurred: {e.message}")
