module github.com/EduOJ/backend

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/duo-labs/webauthn v0.0.0-20200714211715-1daaee874e43
	github.com/fatih/color v1.10.0
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/go-gormigrate/gormigrate/v2 v2.0.0
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-redis/redis/v8 v8.4.11
	github.com/google/uuid v1.2.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/johannesboyne/gofakes3 v0.0.0-20210124080349-901cf567bf01
	github.com/kami-zh/go-capturer v0.0.0-20171211120116-e492ea43421d
	github.com/labstack/echo/v4 v4.2.0
	github.com/labstack/gommon v0.3.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.7.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/minio/md5-simd v1.1.1 // indirect
	github.com/minio/minio-go/v7 v7.0.8
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/xlab/treeprint v1.0.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210220033124-5f55cee0dc0d // indirect
	golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/tools v0.0.0-20200207183749-b753a1ba74fa // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/datatypes v1.0.0
	gorm.io/driver/mysql v1.0.4
	gorm.io/driver/postgres v1.0.8
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.12
)

replace github.com/stretchr/testify v1.6.1 => github.com/leoleoasd/testify v1.6.2-0.20200818074144-885db91dbfe9

replace github.com/johannesboyne/gofakes3 => github.com/leoleoasd/gofakes3 v0.0.0-20210203155129-abef9ae90e02
