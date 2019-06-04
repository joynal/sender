# Push service

### Install packages
```
dep ensure
```

### Run test
```
go test
```

### Deploy

Development:

```
sls deploy -v --force --stage dev
```


Production:

```
sls deploy -v --force --stage prod
```
