module github.com/gptscript-ai/tools/outlook/mail

go 1.23.1

replace (
	github.com/gptscript-ai/tools/outlook/common => ../common
	github.com/hupe1980/golc => github.com/iwilltry42/golc v0.0.113-0.20240802113826-d065a3c5b0c7 // nbformat extension
	github.com/ledongthuc/pdf => github.com/iwilltry42/pdf v0.0.0-20240517145113-99fbaebc5dd3
	github.com/obot-platform/tools/knowledge => ../../knowledge
	github.com/philippgille/chromem-go => github.com/iwilltry42/chromem-go v0.0.0-20250218054308-81ac4c30d459
	github.com/tmc/langchaingo => github.com/StrongMonkey/langchaingo v0.0.0-20240617180437-9af4bee04c8b
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.15.0
	github.com/JohannesKaufmann/html-to-markdown v1.6.0
	github.com/gomarkdown/markdown v0.0.0-20240930133441-72d49d9543d8
	github.com/gptscript-ai/go-gptscript v0.9.6-0.20250331192455-415de950d72d
	github.com/gptscript-ai/tools/outlook/common v0.0.0-20241008222508-3c6174b443e7
	github.com/microsoft/kiota-abstractions-go v1.7.0
	github.com/microsoftgraph/msgraph-sdk-go v1.51.0
	github.com/obot-platform/obot/apiclient v0.0.0-20250417212002-eed4977ade6b
	github.com/obot-platform/tools/knowledge v0.6.9
	github.com/pkoukk/tiktoken-go v0.1.7
)

require (
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

require (
	code.sajari.com/docconv/v2 v2.0.0-pre.4 // indirect
	dario.cat/mergo v1.0.1 // indirect
	github.com/AssemblyAI/assemblyai-go-sdk v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/EndFirstCorp/peekingReader v0.0.0-20171012052444-257fb6f1a1a6 // indirect
	github.com/JalfResi/justext v0.0.0-20221106200834-be571e3e3052 // indirect
	github.com/JohannesKaufmann/dom v0.2.0 // indirect
	github.com/JohannesKaufmann/html-to-markdown/v2 v2.3.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/PuerkitoBio/goquery v1.10.2 // indirect
	github.com/acorn-io/z v0.0.0-20231104012607-4cab1b3ec5e5 // indirect
	github.com/advancedlogic/GoOse v0.0.0-20231203033844-ae6b36caf275 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/aws/aws-sdk-go-v2 v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/textract v1.30.11 // indirect
	github.com/aws/smithy-go v1.20.3 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cjlapao/common-go v0.0.39 // indirect
	github.com/cloudflare/circl v1.6.0 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/ebitengine/purego v0.8.2 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/set v0.2.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gen2brain/go-fitz v1.24.14 // indirect
	github.com/getkin/kin-openapi v0.131.0 // indirect
	github.com/gigawattio/window v0.0.0-20180317192513-0f5467e35573 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.14.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-resty/resty/v2 v2.16.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/hupe1980/go-textractor v0.0.9 // indirect
	github.com/hupe1980/golc v0.0.113 // indirect
	github.com/jaytaylor/html2text v0.0.0-20230321000545-74c2419ad056 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jupiterrider/ffi v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/env v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/ledongthuc/pdf v0.0.0-20240201131950-da5b75280b06 // indirect
	github.com/levigross/exp-html v0.0.0-20120902181939-8df60c69a8f5 // indirect
	github.com/lu4p/cat v0.1.5 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/microcosm-cc/bluemonday v1.0.26 // indirect
	github.com/microsoft/kiota-authentication-azure-go v1.1.0 // indirect
	github.com/microsoft/kiota-http-go v1.4.4 // indirect
	github.com/microsoft/kiota-serialization-form-go v1.0.0 // indirect
	github.com/microsoft/kiota-serialization-json-go v1.0.8 // indirect
	github.com/microsoft/kiota-serialization-multipart-go v1.0.0 // indirect
	github.com/microsoft/kiota-serialization-text-go v1.0.0 // indirect
	github.com/microsoftgraph/msgraph-sdk-go-core v1.2.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/obot-platform/obot/logger v0.0.0-20241217130503-4004a5c69f32 // indirect
	github.com/obot-platform/pdf-parser v0.0.0-20250326062146-23d345e30ecc // indirect
	github.com/olekukonko/tablewriter v0.0.6-0.20230925090304-df64c4bbad77 // indirect
	github.com/otiai10/gosseract/v2 v2.4.1 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.4 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sashabaranov/go-openai v1.38.1 // indirect
	github.com/serpapi/google-search-results-golang v0.0.0-20240325113416-ec93f510648e // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/std-uritemplate/std-uritemplate/go v0.0.57 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tmc/langchaingo v0.1.13 // indirect
	github.com/unidoc/unioffice v1.33.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	gitlab.com/golang-commonmark/html v0.0.0-20191124015941-a22733972181 // indirect
	gitlab.com/golang-commonmark/linkify v0.0.0-20191026162114-a0c2df6c8f82 // indirect
	gitlab.com/golang-commonmark/markdown v0.0.0-20211110145824-bf3e522c626a // indirect
	gitlab.com/golang-commonmark/mdurl v0.0.0-20191124015652-932350d1cb84 // indirect
	gitlab.com/golang-commonmark/puny v0.0.0-20191124015043-9f83538fa04f // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)
