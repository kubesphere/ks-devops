This command line helps to generate `jwtSecret` and `token` for Jenkins client.

## Usage

Run it without Kubernetes:

`go run jwt_cmd.go -s secret`

Run it in Kubernetes, try to find ConfigMap and update it:

`go run jwt_cmd.go`
