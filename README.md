# contracts

The shared wire schemas: protobuf definitions and their generated Go types for the domains a shared host serves, so a caller and the serving host compile against one definition. Domains: anchor, audit, auth, billing, binding, device, domains, enrollment, executive, identity, inference, maintenance, monitoring, notify, options, plugins, provisioning, storage, subject and support.

## Layout

- `protobuf/contracts/<domain>/v1/*.proto`: the sources, in proto package `contracts.<domain>.v1`.
- `<domain>/protobuf/` (and `anchor/pb`, `options`): the generated Go packages, imported as `github.com/bbmumford/contracts/<domain>/protobuf`.

## Use

```sh
go get github.com/bbmumford/contracts
```

```go
import billingpb "github.com/bbmumford/contracts/billing/protobuf"

req := &billingpb.GetCustomerRequest{ /* ... */ }
```

After changing a `.proto` file, regenerate the Go code at the repository root with `buf generate` (`buf.yaml` names the `protobuf` module; `buf.gen.yaml` runs protoc-gen-go and protoc-gen-go-grpc with `module=github.com/bbmumford/contracts`), and commit the sources and the generated files together.

## Consumers

ORBTR library, ai cortex, worker, xyz and most ai endpoints (api, node, worker, app, studio, fleet, console), and io constellation, agent, tools and most io endpoints (node, app, help, mcp, devices).
