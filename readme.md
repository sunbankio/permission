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
    goctl api plugin -plugin permission="-handlerdir /app/adminapi/internal/handler -tpl /dev/tools/plugin/permission/permission.tpl" -api api/zeroapi/adminapi.api
```
### Parameters
| Key | Description |
| -------- | ------- |
| -handlerdir | The directory of go-zero handlers |
| -tpl | Template file to insert in go-zero handler |