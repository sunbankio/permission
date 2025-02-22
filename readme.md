# Permission plugin for go-zero
    - Automatically add permission check to handlers

### Installation
```
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
    goctl api plugin -plugin goctl-docplugin="-handlerdir /app/adminapi/internal/handler -utils github.com/sunbankio/gb-2025/pkg/utils -middlewarePkg github.com/sunbankio/gb-2025/app/adminapi/internal/middleware" -api api/zeroapi/adminapi.api
```
### Parameters
| Key | Description |
| -------- | ------- |
| -handlerdir | The directory of go-zero handlers |