## Candle backend

## Overview
Hack U KOSEN 2023 backend repository

## Requirement
### OS
- MacOS Ventura 13.0
- Arch Linux
### Library
- Go
  - aws-lambda-go
  - aws-sdk-go-v2
  - aws-sdk-go
- TypeScript
  - aws-cdk-lib

 ### Installation
 1. Clone this repository
```
https://github.com/youngeek-0410/candle-backend/
```
2. Install library
```
npm install
```
3. Bootstrapping
```
cdk bootstrap --toolkit-stack-name CandleCDKToolkit --qualifier candle
```
4. Deploy
```
cdk deploy CandleBackendStack --previous-parameters false
```

## Useful commands

* `npm run build`   compile typescript to js
* `npm run watch`   watch for changes and compile
* `npm run test`    perform the jest unit tests
* `npx cdk deploy`  deploy this stack to your default AWS account/region
* `npx cdk diff`    compare deployed stack with current state
* `npx cdk synth`   emits the synthesized CloudFormation template

## Swagger
https://youngeek-0410.github.io/candle-backend/

## Author
- [Yuta Ito](https://github.com/GoRuGoo)
- [Hoku Ishibe](https://github.com/is-hoku)
- [Fumya Sakaguchi](https://github.com/fuu38)
- [Manato Kato](https://github.com/kathmandu777)
