# Sylve Architecture

Sylve consists of several core components, each playing a crucial role in virtualization and storage management.

## Storage

Sylve provides full control over storage management using ZFS, enabling disk operations, dataset creation, and snapshot handling.

### Features:

- **Disk Management:** List all system drives with S.M.A.R.T. data, wipe drives, and initialize with GPT/ZFS.
- **ZFS Management:** Create, list, and delete ZFS pools, datasets, and snapshots.
- **Monitoring & Alerts:** Track pool and dataset statistics, monitor disk health, and receive alerts for potential failures.
- **VM & Jail Storage:** Attach/detach ZFS-backed datastores to VMs and Jails.

### Datastores:

Datastores are created over ZFS datasets and serve as storage for VMs, Jails, and ISOs.

- **Manage Datastores:** Create, list, and delete datastores.
- **Usage Insights:** View statistics and performance data.
- **VM & Jail Integration:** Attach or detach datastores as needed.

## Networking

Networking in Sylve consists of **physical interfaces** and **virtual switches** for flexible connectivity.

### Features:

- **Interface Management:** List, create, and delete network interfaces.
- **Virtual Switches:** Create and manage switches to connect VMs and Jails.
- **VM & Jail Networking:** Attach/detach network interfaces and virtual switches.
- **Monitoring:** Track usage statistics for both interfaces and switches.

## Virtual Machines (VMs)

Sylve leverages **Bhyve** to create and manage virtual machines with extensive configuration options.

### Features:

- **VM Lifecycle:** Create, list, delete, start, stop, and restart VMs.
- **Resource Allocation:** Configure CPU, RAM, disk, and PCI passthrough.
- **Networking:** Attach/detach virtual switches or physical interfaces.
- **Storage:** Attach/detach datastores.
- **Monitoring & Access:** View performance stats, console access, and VNC support.

## Jails

Sylve uses FreeBSD’s **Jails** for lightweight containerization.

### Features:

- **Jail Lifecycle:** Create, list, delete, start, stop, and restart Jails.
- **Resource Allocation:** Configure CPU, RAM, and disk.
- **Networking:** Attach/detach virtual switches or physical interfaces.
- **Storage:** Attach/detach datastores.
- **Monitoring & Access:** View statistics and access the console.

## Development Roadmap

We are implementing features **from left to right** across both APIs and frontend, prioritizing core functionalities before enhancements.

Updated: 10-03-2025
