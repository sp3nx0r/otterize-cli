package mapperclient

//go:generate curl -LJ https://raw.githubusercontent.com/otterize/network-mapper/main/src/mappergraphql/schema.graphql -o ./schema.graphql
//go:generate go run github.com/Khan/genqlient@v0.5.0 ./genqlient.yaml
