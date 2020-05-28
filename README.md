# go-whosonfirst-aws

There are many AWS wrappers. This one is ours.

## Important

This package has been deprecated and will be archived shortly. The functionality it defined has been moved in to individual and discrete packages. They are:

* https://github.com/aaronland/go-aws-ecs
* https://github.com/aaronland/go-aws-lambda
* https://github.com/aaronland/go-aws-s3
* https://github.com/aaronland/go-aws-session

You should use these packages instead of this one.

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This works. Until it doesn't. It has not been properly documented yet.

## DSN strings

```
bucket=BUCKET region={REGION} prefix={PREFIX} credentials={CREDENTIALS}
```

Valid credentials strings are:

* `env:`

* `iam:`

* `{PATH}:{PROFILE}`

## See also

* https://docs.aws.amazon.com/sdk-for-go/

