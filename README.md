# Grafana Dashboard for Admiral Support Tickets

This project provides a Grafana dashboard for visualizing and monitoring the support tickets from Admiral. The dashboard displays various metrics related to support tickets, such as their status, severity, and issue types, with the goal of helping support teams track and analyze the ticket data effectively.

## Features

- **Total Tickets**: Displays the total number of tickets created within a specific time range.
- **Tickets by Severity**: Visualizes the distribution of tickets based on their severity (High, Medium, Low).
- **Tickets by Status**: Displays the count of open and closed tickets.
- **Average Close Time**: Shows the average time taken to close tickets.
- **Unresponded Tickets**: Tracks tickets that have not been responded to after being created.
- **Tickets by Issue Type**: Displays the number of tickets categorized by different issue types.
- **Weekly and Monthly Ticket Counts**: Displays tickets created per week and per month for trend analysis.

## Prerequisites

- **Grafana**: You need a Grafana instance to use this dashboard.
- **Prometheus**: The data for the dashboard is collected via Prometheus metrics.
- **Azure Integration**: Azure data source for fetching support ticket data.
## Grafana Dashboard

The dashboard is designed to provide an overview of the support tickets' status, severity, and issue types. It aggregates data from Azure and presents it in a series of visualizations such as:

- Bar charts for ticket severity distribution.
- Time series graphs for ticket creation over time (weekly, monthly).
- Pie charts for ticket statuses (open/closed).
- Tables for tracking unresponded tickets.
