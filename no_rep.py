import json
with open("source_data.json", "r") as file:
    tickets = json.load(file)
created_partition_keys = []
updated_partition_keys = []
for ticket in tickets:
    event_type = ticket.get("eventType")  
    if event_type == "SupportRequestTicketCreated":
        created_partition_keys.append(ticket.get("partitionKey"))
    elif event_type == "SupportRequestTicketUpdated":
        updated_partition_keys.append(ticket.get("partitionKey"))
no_response_count = 0
for partition_key in created_partition_keys:
    if partition_key not in updated_partition_keys:
        no_response_count += 1
print(f"Number of tickets without a response: {no_response_count}")