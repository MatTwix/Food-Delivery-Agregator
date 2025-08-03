module github.com/MatTwix/Food-Delivery-Agregator/orders-service

go 1.23.6

require github.com/MatTwix/Food-Delivery-Agregator/common v0.0.0-0010101000000-00000000000

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/jackc/pgx/v5 v5.7.5
	github.com/segmentio/kafka-go v0.4.48
	google.golang.org/grpc v1.74.2
)

require (
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)

require (
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/MatTwix/Food-Delivery-Agregator/common => ../common
