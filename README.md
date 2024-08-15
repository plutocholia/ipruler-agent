# Ipruler Agent

## Overview

The ipruler-agent provides several REST APIs that allow you to update the routing configuration of nodes, similar to the functionality offered by Netplan. Currently, it is only compatible with `Linux` machines.

## The Way It Works

After completing the [installation](#installation), a DaemonSet for the [ipruler-agent](https://github.com/plutocholia/ipruler-agent) will be deployed. The ipruler-agent supports two operational modes: `api` and `ConfigBased`.

### `ConfigBased` Mode

In ConfigBased mode, you need to provide the configuration through a configmap, then the ipruler-agent will apply changes after the updated configmap is loaded into the agent container's filesystem and the `CONFIG_RELOAD_DURATION_SECONDS` interval has passed. The agent re-applies the current configuration at each `CONFIG_RELOAD_DURATION_SECONDS` interval, ensuring the nodes' state remains synchronized with the configuration.

You can provide the configuration for the `ConfigBased` mode through the values file in the `ipruler-config` field.

### `api` Mode

In `api` mode, there is an `update` endpoint where you can POST the configuration, which will be applied immediately. Additionally, a separate goroutine will re-apply the last given configuration at each `CONFIG_RELOAD_DURATION_SECONDS` interval, ensuring that the nodes' state remains synchronized with the configuration.

## YAML Configuration Format

The root structure of the configuration file contains four primary sections: `rules`, `settings`, `routes`, and `vlans`.

### `rules`

- **rules**: A list of rules defining the routing table settings.
  - **from**: Specifies the source IP address or network.
  - **table**: Indicates the routing table number to which the rule applies.

### `settings`

- **settings**: Contains additional settings for the configuration.
  - **table-hard-sync**: A list of integers representing routing tables that require hard synchronization. (It will remove any existing routes or rules on the node that do not have a corresponding configuration in the list)

### `routes`

- **routes**: A list of routes specifying the routing details.
  - **to**: The destination IP address or network for the route.
  - **via**: The next-hop IP address through which the route will be directed.
  - **table**: The routing table number to which this route belongs.
  - **dev**: The network device associated with this route.
  - **protocol**: The routing protocol used for this route.
  - **on-link**: A boolean flag indicating whether the route is considered directly connected to the link.
  - **scope**: Specifies the scope of the route (e.g., global, link).

### `vlans`

- **vlans**: A list of VLAN configurations.
  - **name**: The name of the VLAN interface.
  - **link**: The underlying network interface to which the VLAN is attached.
  - **id**: The VLAN ID.
  - **protocol**: The protocol used by the VLAN (e.g., 802.1q or 802.1ad).

### Example YAML Configuration

```yaml
rules:
  - from: 192.168.1.0/24
    table: 100

settings:
  table-hard-sync:
    - 100
    - 200

routes:
  - to: 10.0.0.0/8
    via: 192.168.1.1
    table: 100
    dev: eth0
    protocol: static
    on-link: true
    scope: global

vlans:
  - name: vlan10
    link: eth0
    id: 10
    protocol: 802.1q
```

## Environment Variables

| Environment Variable              | Type   | Default Value          |
|-----------------------------------|--------|------------------------|
| `MODE`                            | string | `api`                  |
| `ENABLE_PERSISTENCE`              | bool   | `false`                |
| `API_PORT`                        | string | `9301`                 |
| `API_BIND_ADDRESS`                | string | `0.0.0.0`              |
| `CONFIG_PATH`                     | string | `./config/config.yaml` |
| `CONFIG_RELOAD_DURATION_SECONDS`  | int    | `15`                   |
| `LOG_LEVEL`                       | string | `INFO`                 |

## Examples

- The content of the [config.yaml](./config/config.yaml) file can be used for both `api` and `ConfigBased` modes.
- [sample valuefile](./charts/ipruler-agent/values.yaml)

## Limitations

## Cleaup Policy

- In `api` mode, you can clean the configuration from the node by sending a POST request to the `cleanup` endpoint.
- In `ConfigBased` mode, to clean the configuration, you need to set the relevant part of the configuration to an empty list.

## Installation

### Helm 

```bash
helm --namespace kube-system upgrade --install \
    --create-namespace --repo https://plutocholia.github.io/ipruler-agent \
    ipruler-agent ipruler-agent --version x.x.x
```

### Default values

| Key | Description | Default Value |
|-----|-------------|---------------|
| agent-config.mode | Specifies the operational mode of the agent. | `api` |
| agent-config.api-port | The port on which the API will be exposed. | `9301` |
| agent-config.config-reload-duration-seconds | Interval in seconds for reapplying the configuration. | `15` |
| agent-config.enable-persistence | Enables or disables persistence of the configuration. | `false` |
| image.repository | Docker repository for the ipruler-agent image. | `plutocholia/ipruler-agent` |
| image.tag | Tag of the Docker image to use. | `~` (chart's app version) |
| image.pullPolicy | Image pull policy for Kubernetes. | `IfNotPresent` |
| resources | Resource limits and requests for the agent container. | `{}` (empty by default) |
| ipruler-config | Configuration for ipruler-agent, including rules and routing tables. |  |
