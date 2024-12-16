
from azure.cosmos import CosmosClient, exceptions
from datetime import datetime, timezone, timedelta
import json

# Fetch sensitive information from environment variables
url = "https://cdbazewp-admiral-core.documents.azure.com:443/"
key = "ZZ5NJmqjAobGMI8w1MHbO4LN4oigpQnM0xwn0q40B9uoHYFqlgRvM6sEhdGv2UxpnJ0nqI1yLRuw2jc5Jx0s4w=="
database_name = "Admiral"
container_name = "Support"

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

        print(f"Ticket details {output_file}")
    else:
        print("No tickets found created")

except exceptions.CosmosHttpResponseError as e:
    print(f"An error occurred: {e.message}")
