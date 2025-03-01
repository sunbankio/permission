# Permission plugin for go-zero
    - Automatically add permission check to handlers

### Installation
```
    go install github.com/sunbankio/permission@latest
    $ go mod tidy
    $ go install
```
## Setup
    -Example Api definition
```
service AdminAPI {
	@doc(
		permission: "user:create1"
	)
	@handler CreateChannel
	post /create (CreateChannelRequest) returns (CreateChannelResponse)
}
```

### Usage
```
    goctl api plugin -plugin permission="-handlerdir /app/adminapi/internal/handler -tpl /dev/tools/plugin/permission/permission.tpl -types github.com/sunbankio/gb-2025/pkg/types/contextkey -dump /app" -api api/zeroapi/adminapi.api


    goctl api plugin -plugin permission="-handlerdir /app/adminapi/internal/handler -tpl /dev/tools/plugin/permission/permission.tpl -types github.com/sunbankio/gb-2025/pkg/types/contextkey -imports github.com/sunbankio/gb-2025/pkg/apierror" -api api/zeroapi/adminapi.api
```
### Parameters
| Key | Description |
| -------- | ------- |
| -handlerdir | The directory of go-zero handlers |
| -tpl | Template file to insert in go-zero handler |
| -types | Types package where the contextkey is defined |
| -dump  | Dump all permissions detected |
| -imports | custom imports to be appended on the handler |


git tag v1.0.5
git push origin v1.0.5
GOPROXY=proxy.golang.org go list -m github.com/sunbankio/permission@v1.0.5