# DEAR WMS Integration

https://dearsystems.com.cn

**DEAR WMS Integration** is a Go application that connects DEAR Inventory with Warehouse Management System (WMS). This integration automates inventory management tasks like stock transfers, purchase stock authorizations, and product availability checks, enabling smooth, real-time synchronization of stock levels.

## Features

- **Product Availability Check**: Fetches current stock data from DEAR Inventory.
- **Stock Transfers**: Automatically updates stock levels in response to transfer events.
- **Purchase Stock Authorization**: Syncs stock after purchase orders are authorized.
- **Scheduled Inventory Check**: Runs daily cron jobs to verify and reconcile stock discrepancies between DEAR and WMS.
- **Webhook Support**: Supports multiple webhooks for real-time data updates.

## Table of Contents

- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [CI/CD](#ci-cd)
- [License](#license)

## Getting Started

### Prerequisites

- **Go**: version 1.17 or later
- **Docker**: to run the application in a containerized environment
- **Viper**: for configuration management
- **Echo**: for handling HTTP routes
- **Cron**: for scheduling tasks

### Installation

Clone the repository:

```bash
git clone https://github.com/JLGUOZIN/dear-wms-integration.git
cd dear-wms-integration
