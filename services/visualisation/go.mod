module github.com/Bad-Utya/myforebears-backend/services/visualisation

go 1.25.1

require (
	github.com/Bad-Utya/myforebears-backend/gen/go v0.0.0-20260302235630-1f4dfc9b1b21
	github.com/Bad-Utya/myforebears-backend/utility v0.0.0
	github.com/aws/aws-sdk-go-v2/config v1.31.13
	github.com/aws/aws-sdk-go-v2/credentials v1.18.17
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.5
	github.com/google/uuid v1.6.0
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.79.1
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.7 // indirect
	github.com/aws/smithy-go v1.23.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/Bad-Utya/myforebears-backend/utility => ../../utility

replace github.com/Bad-Utya/myforebears-backend/gen/go => ../../gen/go
