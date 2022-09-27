module github.com/signaux-faibles/keycloakUpdater/v2

go 1.18

//replace github.com/signaux-faibles/libwekan => /home/christophe/Project/libwekan
//replace github.com/signaux-faibles/libwekan => /Users/raphaelsquelbut/devhome/workspace/signaux-faibles/libwekan
replace github.com/signaux-faibles/libwekan => ../libwekan

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/Nerzal/gocloak/v11 v11.0.2
	github.com/cnf/structhash v0.0.0-20201127153200-e1b16c1ebc08
	github.com/mattn/go-colorable v0.1.12
	github.com/ory/dockertest/v3 v3.9.1
	github.com/pkg/errors v0.9.1
	github.com/signaux-faibles/libwekan v0.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/snowzach/rotatefilehook v0.0.0-20220211133110-53752135082d
	github.com/snowzach/writerhook v0.0.0-20180327172656-2eef47d6a0bf
	github.com/stretchr/testify v1.8.0
	github.com/tealeg/xlsx/v3 v3.2.4
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.14+incompatible // indirect
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/go-resty/resty/v2 v2.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.1.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/shabbyrobe/xmlwriter v0.0.0-20200208144257-9fca06d00ffa // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.10.2
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
